package cmd

import (
	"fmt"

	"github.com/melih-ucgun/monarch/internal/crypto"
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Şifreleme ve anahtar yönetimi işlemlerini yapar",
}

var genKeyCmd = &cobra.Command{
	Use:   "gen-key",
	Short: "Yeni bir age anahtar çifti oluşturur",
	Run: func(cmd *cobra.Command, args []string) {
		priv, pub, err := crypto.GenerateKey()
		if err != nil {
			fmt.Printf("❌ Anahtar oluşturulamadı: %v\n", err)
			return
		}
		fmt.Printf("🗝️  Private Key (Bunu güvenli saklayın!): %s\n", priv)
		fmt.Printf("📢 Public Key (Şifrelemek için kullanın): %s\n", pub)
	},
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt [metin]",
	Short: "Bir metni verilen public key ile şifreler",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pubKey, _ := cmd.Flags().GetString("key")
		if pubKey == "" {
			fmt.Println("❌ Lütfen --key bayrağı ile bir public key belirtin.")
			return
		}

		encrypted, err := crypto.Encrypt(args[0], pubKey)
		if err != nil {
			fmt.Printf("❌ Şifreleme hatası: %v\n", err)
			return
		}

		fmt.Println("🔒 Şifreli Metin (Bunu YAML dosyasına yapıştırın):")
		fmt.Println(encrypted)
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.AddCommand(genKeyCmd)
	secretsCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringP("key", "k", "", "Şifreleme için kullanılacak public key")
}
