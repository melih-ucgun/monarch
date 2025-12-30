package core

import "fmt"

// StateUpdater interface'i, Engine'in state paketine doğrudan bağımlı olmamasını sağlar.
type StateUpdater interface {
	UpdateResource(resType, name, targetState, status string) error
}

// ConfigItem, motorun işleyeceği ham konfigürasyon parçasıdır.
type ConfigItem struct {
	Name   string
	Type   string
	State  string
	Params map[string]interface{}
}

// Engine, kaynakları yöneten ana yapıdır.
type Engine struct {
	Context      *SystemContext
	StateUpdater StateUpdater // Opsiyonel: State yöneticisi
}

// NewEngine yeni bir motor örneği oluşturur.
func NewEngine(ctx *SystemContext, updater StateUpdater) *Engine {
	return &Engine{
		Context:      ctx,
		StateUpdater: updater,
	}
}

// ResourceCreator fonksiyon tipi
type ResourceCreator func(resType, name string, params map[string]interface{}, ctx *SystemContext) (ApplyableResource, error)

// ApplyableResource arayüzü
type ApplyableResource interface {
	Apply(ctx *SystemContext) (Result, error)
	GetName() string
}

// Run, verilen konfigürasyon listesini işler.
func (e *Engine) Run(items []ConfigItem, createFn ResourceCreator) error {
	errCount := 0

	for _, item := range items {
		// Params hazırlığı
		if item.Params == nil {
			item.Params = make(map[string]interface{})
		}
		item.Params["state"] = item.State

		// 1. Kaynağı oluştur
		res, err := createFn(item.Type, item.Name, item.Params, e.Context)
		if err != nil {
			Failure(err, "Skipping invalid resource definition: "+item.Name)
			errCount++
			continue
		}

		// 2. Kaynağı uygula
		result, err := res.Apply(e.Context)

		status := "success"
		if err != nil {
			status = "failed"
			errCount++
			fmt.Printf("❌ [%s] Failed: %v\n", item.Name, err)
		} else if result.Changed {
			fmt.Printf("✅ [%s] %s\n", item.Name, result.Message)
		} else {
			fmt.Printf("ℹ️  [%s] OK\n", item.Name)
		}

		// 3. Durumu Kaydet (Eğer DryRun değilse)
		if !e.Context.DryRun && e.StateUpdater != nil {
			// Başarısız olsa bile son deneme durumunu "failed" olarak kaydediyoruz
			saveErr := e.StateUpdater.UpdateResource(item.Type, item.Name, item.State, status)
			if saveErr != nil {
				fmt.Printf("⚠️ Warning: Failed to save state for %s: %v\n", item.Name, saveErr)
			}
		}
	}

	if errCount > 0 {
		return fmt.Errorf("encountered %d errors during execution", errCount)
	}
	return nil
}
