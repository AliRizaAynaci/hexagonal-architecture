# Hexagonal Architecture: Nedir, Neden Önemlidir, Go ile class'lar olmadan nasıl kurgulanır?

Mülakatlara hazırlanırken birçok konuyu PoC (proof of concept) olarak uygulamaya çalışıyorum fakat 
şu ana kadar çalıştığım konular içerisinde beni en çok zorlayan ve sıfırdan tasarlayıp kurması 
gerçekten zor hissettiren bir konu olduğundan ileride tekrar okuyup hatırlamamı kolaylaştırmak 
adına aslında bu yazıyı yazıyorum. Bu yüzden daha çok çalışma notları gibi olacaktır, 
umarım sizlere de faydalı olur.

## Table of Contents
- [Neden Hexagonal Architecture?](#neden-hexagonal-architecture)
- [Peki Hexagonal Architecture Dediğimiz Nedir?](#peki-hexagonal-architecture-dedigimiz-nedir)
- [Hexagonal Architecture'ı bir senaryo üzerinde sıfırdan kuralım](#hexagonal-architecturei-bir-senaryo-uzerinde-sifirdan-kuralim)
    - [1. Adım: Domain katmanını oluşturma](#1-adim-domain-katmanini-olusturma)
    - [2. Adım: Ports](#2-adim-ports)
    - [3. Adım: Application - Servis Katmanı](#3-adim-application---servis-katmani)
    - [4. Adım: Adapters (Driven - Çıkış Adaptörleri)](#4-adim-adapters-driven---cikis-adaptorleri)
    - [5. Adım: Driving Adapters (HTTP Handler)](#5-adim-driving-adapters-http-handler)
    - [6. Adım: Main (Composition Root & Dependency Injection)](#6-adim-main-composition-root--dependency-injection)
- [Demo](#demo)


## Neden Hexagonal Architecture?

Klasik proje geliştirme süreçlerimde aslında çokça karşılaştığım bir mimari değil. 
Fakat, production seviyesinde bir yazılım projesi geliştirirken projenin çalışıyor olması ve işi yapıyor 
olması yeterli olmuyor. Proje bittikten ve teslim edilip production'a çıktıktan sonra maintain edilebilir 
olması gerekiyor, ki projelerin en çok maliyet çıkartan aşaması da bakım aşamasıdır. 
Bu açıdan baktığımızda bir yazılım mühendisinin özellikle günümüz AI çağında kod yazmayı değil, 
bu kodun ileride başka mühendisler tarafından kullanılacağını ve bakım aşamasında update'ler, 
yeni özellikler ekleneceğini düşünerek projeyi tasarlamalıdır. 
İşte Hexagonal Architecture tam olarak bu noktada çok büyük bir önem kazanıyor.


## Peki Hexagonal Architecture Dediğimiz Nedir?

Hexagonal Architecture en basit tabiriyle yüksek seviyeli sınıfların (Domain, Entities vb.), 
düşük seviyeli sınıflara (DB, LOG, External APIs vb.) bağlı olmamasıdır. 
Yani buradaki mantık SOLID (Single Responsibility, Open-Closed, Liskov Substitution, Interface Segregation, 
Dependency Inversion) prensipleriyle birebir uyum içerisindedir.

Buradaki mimariyi Port-Adapter olarak düşünelim;
Örnek verecek olursak, sistemimize gelen veriyi kaydetmek istiyoruz. 
Hexagonal Architecture'da 'Veriyi Kaydet' diye bir kural koyarız (Port), ama bunu gerçekten kimin yaptığını 
(Postgres mi, Mongo mu?) domain bilmez, bilmesine gerek de kalmaz. Böylece bu parçaları değiştirmek, 
yani sistemi maintain etmek çok daha basitleşir. İleride sistemimizin kullandığı DB'yi değiştirmek istesek 
oturup business katmanını yeniden yazmamıza gerek kalmadan, sadece önceden kurgulamış olduğumuz 
"Port"a uygun yeni bir Adapter yazmak kadar kolay olacaktır.


## Hexagonal Architecture'ı bir senaryo üzerinde sıfırdan kuralım

Basit bir senaryo olacak, kısaca yapmak istediğimiz senaryo **Online Konser Bilet Satış Sistemi** olsun.

### 1.Adim Domain katmanini olusturma

Mimariyi kurmaya en içten, yani domain katmanından başlıyoruz. 
Hexagonal Architecture'da altın kural şudur: **Bağımlılıklar her zaman içe doğru bakar.** 
Bu yüzden dış dünyayı (Database, Web Servers, Message Queues) düşünmeden önce, 
uygulamanın ne iş yapacağını ve kurallarını tanımlamalıyız.

Burada HTTP, SQL, JSON veya herhangi bir kütüphane bağımlılığı göremezsiniz. 
Sadece saf Go struct'ları ve iş mantığı (Business Logic) metodları yer alır.

Senaryomuzdaki en temel varlık (Entity) olan Concert yapısını tasarlayarak başlayalım.
Bir konserin adı, kapasitesi ve satılan bilet sayısı gibi özellikleri vardır. 
Ayrıca "Bilet satılabilir mi?" gibi kontrolleri de burada yaparız.

**Dosya:** `internal/core/domain/concert.go`
```go
package domain

import (
	"errors"
	"time"
)

// Dikkat: Burada `json:"..."` veya `db:"..."` tagleri YOK.
// Çünkü Domain, veritabanından veya HTTP'den habersiz olmalı!
// Yarın öbür gün JSON yerine gRPC kullanırsak burası değişmemeli
type Concert struct {
	ID          string
	Name        string
	Capacity    int
	SoldTickets int
	Date        time.Time
}

// Service katmani hatanin ne oldugunu anlayip ona gore HTTP 409 vs donmesi icin
var ErrCapacityExceeded = errors.New("concert capacity exceeded")


// NewConcert - Bir Constructor
// Geçerli bir konser nesnesi oluşturmak için kuralları buraya koyarız.
// Örneğin: Kapasite negatif olamaz kontrolü buraya eklenebilir.
func NewConcert(id string, name string, capacity int, date time.Time) *Concert {
    return &Concert{
        ID:          id,
        Name:        name,
        Capacity:    capacity,
        SoldTickets: 0,
        Date:        date,
    }
}

// CanSell - Business Rule buradadır.
// Database'e bakmaz, sadece elindeki veriye göre karar verir.
func (c *Concert) CanSell(quantity int) error {
    if c.SoldTickets+quantity > c.Capacity {
        return ErrCapacityExceeded
    }
    return nil
}

func (c *Concert) Sell(quantity int) error {
    if c.SoldTickets+quantity > c.Capacity {
        return ErrCapacityExceeded
    }
    c.SoldTickets += quantity
    return nil
}
```

### 2. Adım: Ports

Domain nesnemizi oluşturduk ama şu an tek başına duruyor, dış dünya ile bir iletişimi bulunmuyor. 
Şimdi bu nesneye giriş ve çıkış Portları inşa edeceğiz. Bu portları interface olarak tanımlıyoruz.

Bizim senaryomuzda bu portlar:

1. **Giriş Portu (Service Interface):** Dışarıdan gelen isteklerin 
(örneğin bir HTTP isteği) iş mantığımıza ulaşmasını sağlar.
Kodumuzda: `ConcertService` (Bilet al, Konser oluştur emirlerini buradan alırız).

2. **Çıkış Portu (Repository Interface):**
İşlenen veriyi kalıcı hale getirmek (Save/Get) için kullanacağımız port.

Buradaki en kritik nokta şudur: Kodumuzda "Postgres" veya "MySQL" kelimesi geçmez.
Sadece "Repository" deriz. Arkada hangi teknolojinin olduğu Core katmanını ilgilendirmez.

![Başlıksız Diyagram.drawio.png](Ba%C5%9Fl%C4%B1ks%C4%B1z%20Diyagram.drawio.png)

**Dosya:** `internal/core/ports/ports.go`
```go
package ports

import "hexagonal/internal/core/domain"

// Driven Port (Secondary): Veritabanı işlemleri için gerekli sözleşme.
// Domain katmanı veritabanı detayını bilmez, sadece bu metodları bilir.
// İleride buraya Postgres, Mongo veya In-Memory ne bağlarsan bağla, Domain değişmez.
type ConcertRepository interface {
    Get(id string) (*domain.Concert, error)
    Save(concert domain.Concert) error
}

// Driving Port (Primary): Dış dünyanın (HTTP, CLI, gRPC) kullanacağı servis sözleşmesi.
// Service katmanı bu interface'i implemente edecek.
type ConcertService interface {
    CreateConcert(id string, name string, capacity int) (*domain.Concert, error)
    BuyTicket(concertID string, quantity int) error
}
```

### 3. Adım: Application - Servis Katmanı

Entity'miz hazır, portlarımız (Interface) hazır.
Şimdi bu portları birbirine bağlayacak olan İş Akışını (Business Flow) yazalım.

**Servis katmanının görevi:**
1. Dış dünyadan gelen emri al (`ConcertService` interface'ini implemente eder).
2. Gerekli iş kurallarını işletir (`Concert` entity'si üzerindeki metodları çağırır).
3. Sonucu dış dünyaya kaydeder (`ConcertRepository` interface'ini kullanır).

Buradaki önemli detay **Dependency Injection** (Bağımlılık Enjeksiyonu) yapmamızdır. 
Servisimiz çalışmak için bir Repository'ye ihtiyaç duyar, ama bu Repository'nin ne olduğunu 
(Postgres, MySQL, Memory) bilmez. Sadece interface'i bilir.

![Başlıksız Diyagram.drawio (1).png](Ba%C5%9Fl%C4%B1ks%C4%B1z%20Diyagram.drawio%20%281%29.png)

**Dosya:** `internal/core/services/concert_service.go`
```go
package services

import (
	"hexagonal/internal/core/domain"
	"hexagonal/internal/core/ports"
	"errors"
	"time"
)

// service - ConcertService portunu implemente eden yapımız.
// Küçük harfle başlattık çünkü dışarıdan direkt erişilmesin, Constructor ile verilsin.
type service struct {
	repo      ports.ConcertRepository
}

// NewConcertService - Servis oluşturmak için kullanılan Constructor fonksiyonu.
// Bu fonksiyon herhangi bir `Repository` alir ve bize bir ports.ConcertService döner.
func NewConcertService(r ports.ConcertRepository) ports.ConcertService {
	return &service{
		repo:      r,
	}
}

// CreateConcert - İş Akışı
func (s *service) CreateConcert(id string, name string, capacity int) (*domain.Concert, error) {
	// Saf Domain nesnesini oluştur
	newConcert := domain.NewConcert(id, name, capacity, time.Now().AddDate(0, 1, 0))

	// Repository Port üzerinden kaydet
	if err := s.repo.Save(*newConcert); err != nil {
		return nil, err
	}

	return newConcert, nil
}

func (s *service) BuyTicket(concertID string, quantity int) error {
	// Repository Port uzerinden veriyi cek
	concert, err := s.repo.Get(concertID)
	if err != nil {
		return err
	}
	if concert == nil {
		return errors.New("concert not found")
	}

	// Domain Logicte bulunan kurallari uygula
	if err := concert.CanSell(quantity); err != nil {
		return err
	}

	// Verinin durumunu guncelle
	concert.Sell(quantity)

	// Repository Port araciligi ile kaydet
	if err := s.repo.Save(*concert); err != nil {
		return err
	}

	return nil
}
```

### 4. Adım: Adapters

Core katmanımızı tamamladık, fakat hala dış dünya ile iletişime geçmiyor uygulamamız. 
Şimdi sıra soyut dünyayı dış dünyaya bağlamakta. Hexagonal Architecture'da, Interface'leri (Portları) 
implemente eden sınıflara "Adapter" diyoruz.

Şimdi sıra geldi bu kadar uğraşın sonucundaki en faydalı kısmı anlamaya: Core katmanımız `ConcertRepository` diye bir interface bekliyor. Biz ona şu an basit bir **In-Memory** veritabanı vereceğiz. 
Yarın öbür gün bunu **PostgresAdapter** ile değiştirdiğimizde Core katmanının ruhu bile duymayacak.

İlk adaptörümüzü, yani veritabanı katmanını yazalım. 
Şimdilik Docker veya SQL ile uğraşmamak için veriyi RAM'de tutan bir yapı kuruyoruz.

**Dosya:** `internal/adapters/repository/memory_repo.go`

```go
package repository

import (
	"errors"
	"hexagonal/internal/core/domain"
	"sync"
)

// InMemoryRepository - Veritabanı yerine geçecek geçici yapı.
// Neden sync.RWMutex var? -> Eşzamanlı (Concurrent) isteklerde map patlamasın diye.
type InMemoryRepository struct {
	db map[string]domain.Concert
	mu sync.RWMutex // Read Lock icin RWMutex kullanilmali
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		db: make(map[string]domain.Concert),
	}
}

// Get - veriyi okuma islemi
// Read Lock (RLock) kullanıyoruz çünkü okurken başkaları da okuyabilir
// sadece yazan olmasın yeter.
func (r *InMemoryRepository) Get(id string) (*domain.Concert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	concert, ok := r.db[id]
	if !ok {
		return nil, errors.New("concert not found")
	}

	return &concert, nil
}

// Mutual Exclusion
// Write Lock (Lock) kullanıyoruz. Biz yazarken kimse okuyamaz ve yazamaz 
func (r *InMemoryRepository) Save(concert domain.Concert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.db[concert.ID] = concert
	return nil
}
```

Gördüğümüz gibi, bu dosya `internal/adapters` klasöründe yaşıyor. 
Core katmanındaki ports paketini import ediyor ve oradaki kurallara uyuyor. 
Bu sayede Core katmanı, verinin bir map içinde mi yoksa 
Postgres tablosunda mı tutulduğunu asla bilmiyor.

### 5. Adım: Driving Adapters (HTTP Handler)
Sistemimiz neredeyse hazır. Veriyi tutabiliyoruz, işleyebiliyoruz ama henüz dışarıdan kimse bu sistemi 
tetikleyemiyor. Şimdi Fiber framework'ünü kullanarak bir HTTP REST API yazacağız.

Bu katmanın (Adapter) tek bir görevi vardır: Gelen HTTP isteğini (JSON) okumak ve Service katmanındaki 
ilgili fonksiyonu çağırmak. İş mantığı (Business Logic) kesinlikle burada olmaz. 
Burası sadece bir "Tercüman" gibidir.

**Dosya:** `internal/adapters/handler/http_handler.go`

```go
package handler

import (
	"hexagonal/internal/core/ports"
	"github.com/gofiber/fiber/v2"
)

// HTTPHandler - Web isteklerini karşılayan adaptör.
// Service katmanını çağırır, JSON verisini parse eder.
type HTTPHandler struct {
	svc ports.ConcertService
}

func NewHTTPHandler(svc ports.ConcertService) *HTTPHandler {
	return &HTTPHandler{
		svc: svc,
	}
}

func (h *HTTPHandler) CreateConcert(c *fiber.Ctx) error {
	// İstekten gelen JSON body için bir DTO (Data Transfer Object) tanımladık.
	// Neden: Domain nesnesini doğrudan dışarı açmak istemeyiz (Security & Decoupling).
	var body struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Capacity int    `json:"capacity"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid bodu"})
	}

	// Service katmanini cagir
	concert, err := h.svc.CreateConcert(body.ID, body.Name, body.Capacity)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(concert)
}

func (h *HTTPHandler) BuyTicket(c *fiber.Ctx) error {
	var body struct {
		ConcertID string `json:"concert_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	err := h.svc.BuyTicket(body.ConcertID, body.Quantity)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(fiber.Map{"message": "ticket bought successfully"})
}
```
Bu kodla birlikte uygulamanın Dış dünya ile etkileşimini de tamamlamış olduk. 
Artık elimizde tam teşekküllü bir Hexagonal yapı var. Sadece `main.go` dosyasında **Dependency Injection** 
işlemlerini yapmak kaldı.

Kod tarafında Core, Ports ve Adapters katmanlarını tamamladık. 
Soyut kalan bu kavramların bir araya geldiğinde nasıl bir veri akışı oluşturduğunu görelim, 
yapıyı kafamızda oturtmak için çok önemli.


![Başlıksız Diyagram.drawio (2).png](Ba%C5%9Fl%C4%B1ks%C4%B1z%20Diyagram.drawio%20%282%29.png)

### 6. Adım: Main (Composition Root & Dependency Injection)

Şimdiye kadar yazdığımız tüm parçalar (Domain, Service, Repository, Handler) 
birbirinden bağımsız duruyor. Hiçbiri kendi başına çalışamaz. `main.go` dosyası, 
bu parçaların birbirine bağlandığı, yani **Dependency Injection** işleminin yapıldığı yerdir.

Bu dosyanın görevi çok basit:
1. **Adaptörleri Oluştur:** Somut nesneleri (MemoryRepository, Fiber App) yarat.
2. Servisleri Oluştur ve Enjekte Et: Servisi oluştururken ona "Al, kullanacağın repo bu" de.
3. Uygulamayı Başlat: Web sunucusunu ayağa kaldır.

**Dosya:** `cmd/api/main.go`
```go
package main

import (
	"hexagonal/internal/adapters/handler"
	"hexagonal/internal/adapters/repository"
	"hexagonal/internal/core/services"
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// ADAPTERS: Önce bağımlılıkları oluşturuyoruz.
	// Şu an MemoryRepo kullanıyoruz. 
	// Yarın PostgresRepo'ya geçersek sadece burayı değiştireceğiz.
	repo := repository.NewMemoryRepository()

	// CORE - Service: Servisi oluştururken Repoyu içine enjekte ediyoruz (DI).
	svc := services.NewConcertService(repo)

	// ADAPTERS: Handler'ı oluştururken Servisi içine enjekte ediyoruz.
	h := handler.NewConcertHandler(svc)

	// SERVER: Web sunucusunu ayağa kaldırıyoruz.
	app := fiber.New()

	// Routelari tanımlıyoruz
	app.Post("/concerts", h.CreateConcert)
	app.Post("/concerts/:id/buy", h.BuyTicket)

	log.Println("Server running on :3000")
	log.Fatal(app.Listen(":3000"))
}
```


### Demo

Kodlarımızı yazdık, parçaları birleştirdik. 
Şimdi terminalden uygulamamızı ayağa kaldıralım ve sonuçları gözlemleyelim.

Aşağıdaki çıktılar, bu projenin temel mantığının üzerine Zap Logger ve Kafka (Event Streaming) 
gibi yapıların eklenmiş halidir. Temel mimari (Hexagonal) birebir aynı kalmakla birlikte, 
Adaptör katmanında yapılan geliştirmelerin Core katmanını nasıl etkilemediğini 
veya etkilememesi gerektiğini canlı olarak görmüş oluyoruz.

#### Önce bir konser oluşturalım:

```shell
curl -X POST http://localhost:3000/concerts \
     -H "Content-Type: application/json" \
     -d '{"id": "istanbul-konser", "name": "Istanbul Festival", "capacity": 50000}'
```
API Response:
```shell
ID          : istanbul-konser
Name        : Istanbul Festival
Capacity    : 50000
SoldTickets : 0
Date        : 2026-03-09T03:00:38.697811+03:00
```
Uygulama Logları:

```

{
    "level":"info",
    "timestamp":"2026-02-09T03:00:38.697+0300",
    "caller":"services/concert_service.go:40",
    "msg":"concert created successfully",
    "pid":1472,
    "concert_id":
    "istanbul-konser",
    "name":"Istanbul Festival",
    "capacity":50000
}
```

#### Şimdi de bu konsere bilet satın alalım:
```shell
curl -X POST http://localhost:3000/tickets \
     -H "Content-Type: application/json" \
     -d '{"concert_id": "istanbul-konser", "quantity": 2}'
```
API Response
```shell
{
    "message": "ticket bought successfully"
}
```

Uygulama Logları:

```

{
    "level":"info",
    "timestamp":"2026-02-09T03:02:16.073+0300",
    "caller":"services/concert_service.go:92",
    "msg":"ticket sold and event published",
    "pid":1472,
    "event_id":"01345587-ac36-454a-94a1-5af0af909112",
    "concert_id":"istanbul-konser"
}
```
Gördüğümüz gibi, bir bilet satıldığında sistem sadece veritabanını güncellemekle kalmadı, 
aynı zamanda bir domain event fırlattı. GitHub reposunu incelediğinizde göreceksiniz ki, 
Hexagonal Architecture sayesinde bu karmaşıklığı yönetmek oldukça kolay.

Bu projenin kaynak kodlarına, docker-compose dosyasına ve Kafka entegrasyonunun 
tam haline aşağıdaki Github reposundan ulaşabilirsiniz:

https://github.com/AliRizaAynaci/hexagonal-architecture