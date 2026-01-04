# Veto Projesi İlerleme Raporu (Ocak 2026)

Veto projesi, altyapı yönetimi ve "fleet management" (filo yönetimi) kabiliyetlerini artıran kritik güncellemeleri başarıyla tamamlamıştır. Mevcut durum itibarıyla sistem, hem tekil makinelerde hem de uzak sunucu kümelerinde güvenli ve şeffaf bir yapılandırma yönetimi sunmaktadır.

## Tamamlanan Temel Özellikler

### 1. Gelişmiş "Prune" (Yönetilmeyen Kaynak Tespiti)
Sistem artık sadece yapılandırmada olanları kurmakla kalmıyor, yapılandırmada **olmayan** gereksiz kaynakları da tespit edebiliyor:
- **Global Budama**: Tüm ana paket yöneticileri (Apt, DNF, Pacman, Brew, Apk, Zypper, Yum) için yüklü paket listeleme desteği eklendi.
- **Kapsamlı (Scoped) Budama**: Belirli dizinler (örn: `/etc/ssh/conf.d`) için `prune: true` parametresi ile yönetilmeyen dosyaların otomatik temizliği sağlandı.
- **Güvenlik**: Yıkıcı işlemler öncesinde detaylı önizleme ve kullanıcı onayı mekanizmaları güçlendirildi.

### 2. Gözlemlenebilirlik ve Yapılandırılmış Loglama
- `slog` entegrasyonu ile tüm işlemler makineler arası takip edilebilir hale geldi.
- **Host Bağlamlı Loglar**: Çoklu makine (fleet) modunda, her log satırı hangi makineden geldiğini (`host="xyz"`) belirtecek şekilde etiketlendi.
- **Verimlilik**: `-v`, `-vv`, `-vvv` bayrakları ile log seviyesi dinamik olarak ayarlanabilir hale getirildi.

### 3. Çoklu Makine (Fleet) Yönetimi
- SSH transport katmanı ve `FleetManager` aracılığıyla onlarca makineye paralel yapılandırma uygulama desteği stabilize edildi.

---

## Geliştirilmesi Gereken Noktalar ve Yol Haritası

Projenin bir sonraki aşamasında odaklanılması önerilen noktalar şunlardır:

### 1. Bağımlılık Yönetimi Geliştirmeleri (Dependency Graph)
- Şu anki yapı katmanlı (layer-based) paralel çalışmaktadır. Daha karmaşık senaryolar için kaynaklar arası **DAG (Directed Acyclic Graph)** tabanlı bir bağımlılık çözücüye geçilmesi performansı artıracaktır.

### 2. İleri Seviye Güvenlik (Secrets Management)
- Vault entegrasyonu veya şifrelenmiş değişkenlerin (secrets) `runtime` sırasında daha güvenli (bellekte şifreli tutularak) işlenmesi sağlanabilir.

### 3. Durum Kontrolü (Drift Detection) İyileştirmesi
- Yapılandırmanın sadece uygulanması değil, zaman içinde değişip değişmediğini (drift) raporlayan bir `veto audit` veya `veto drift-check` komutu eklenebilir.

### 4. Ekosistem Desteği
- **Container Desteği**: Docker/Podman imajları için özel bir adapter geliştirilebilir.
- **Snap/Flatpak Desteği**: Mevcut paket yöneticilerine ek olarak container tabanlı paketleme sistemleri için `Lister` desteği eklenebilir.

### 5. Web Dashboard (Opsiyonel)
- Fleet modunda çalışan operasyonların durumunu görselleştiren, Prometheus/Grafana ile entegre bir dashboard arayüzü eklenebilir.

> [!NOTE]
> Proje şu anki haliyle prodüksiyon seviyesinde temel altyapı otomasyonu için oldukça stabil ve güvenli bir zemin sunmaktadır.
