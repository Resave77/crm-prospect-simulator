# Google Maps / Places API Pricing Guide

Dokumen ini merangkum API Google yang digunakan aplikasi berdasarkan implementasi saat ini.

> Harga menggunakan pricing global tier pertama Google dalam USD. Harga aktual dapat berbeda berdasarkan region, volume, dan konfigurasi billing project. Terakhir diverifikasi: 24 Agustus 2026.

## Ringkasan API dan Pricing

| API yang digunakan | Endpoint / fungsi | SKU billing | Harga per 1.000 hit | Estimasi per hit | Free tier |
|---|---|---|---:|---:|---:|
| Nearby Search (New) | `POST /v1/places:searchNearby` | Nearby Search Pro | $32 | $0,032 | 5.000 hit/bulan |
| Text Search (New) | `POST /v1/places:searchText` | Text Search Pro | $32 | $0,032 | 5.000 hit/bulan |
| Place Details - Core | `GET /v1/places/{placeId}` | Place Details Pro | $17 | $0,017 | 5.000 hit/bulan |
| Place Details - Business Info | `GET /v1/places/{placeId}` | Place Details Enterprise | $20 | $0,020 | 1.000 hit/bulan |
| Place Details - Full Prospect Finder | `GET /v1/places/{placeId}` | Place Details Enterprise + Atmosphere | $25 | $0,025 | 1.000 hit/bulan |
| Place Photo (New) | `GET /v1/{photoName}/media` | Place Details Photos | $7 | $0,007 | 1.000 hit/bulan |
| Custom Search JSON API - Image Search | `GET https://www.googleapis.com/customsearch/v1` dengan `searchType=image` | Custom Search JSON API | $5 | $0,005 | 100 query/hari, sekitar 3.000 query/30 hari |

## Tabel API, Field, Harga, dan Rata-rata Hit per Flow

Flow standar yang digunakan untuk kolom **Rata-rata hit per flow** adalah:

```text
1. User melakukan search
2. User memilih satu tempat
3. Aplikasi membuka detail tempat
4. Aplikasi menampilkan satu foto pertama jika tersedia
```

Dengan asumsi cache MISS, satu flow standar biasanya memakai **3 hit provider**: 1 hit search, 1 hit detail, dan 1 hit photo. Custom Search Image hanya dipanggil jika user meminta pencarian foto menu.

| API | Field yang diambil | Kegunaan di aplikasi | SKU | Harga per hit | Free tier per bulan | Rata-rata hit per flow |
|---|---|---|---|---:|---:|---:|
| Nearby Search (New) | `id`<br>`displayName`<br>`formattedAddress`<br>`primaryTypeDisplayName`<br>`types`<br>`businessStatus`<br>`location` | Google Place ID<br>Nama bisnis/tempat<br>Alamat tampilan<br>Kategori utama<br>Tipe/kategori Google<br>Status bisnis<br>Latitude dan longitude | Nearby Search Pro | $0,032 | 5.000 hit | 1 hit jika search berdasarkan lokasi/kategori |
| Text Search (New) | `id`<br>`displayName`<br>`formattedAddress`<br>`primaryTypeDisplayName`<br>`types`<br>`businessStatus`<br>`location` | Google Place ID<br>Nama bisnis/tempat<br>Alamat tampilan<br>Kategori utama<br>Tipe/kategori Google<br>Status bisnis<br>Latitude dan longitude | Text Search Pro | $0,032 | 5.000 hit | 1 hit jika search berdasarkan keyword |
| Place Details - Core | Semua field search di atas<br>`googleMapsUri`<br>`photos` | Detail dasar tempat<br>Link ke Google Maps<br>Metadata foto tersedia | Place Details Pro | $0,017 | 5.000 hit | 0 hit pada flow Prospect Finder utama; 1 hit jika halaman memakai Core Detail |
| Place Details - Business Info | Semua field detail dasar<br>`nationalPhoneNumber`<br>`internationalPhoneNumber`<br>`websiteUri`<br>`rating`<br>`userRatingCount`<br>`regularOpeningHours` | Informasi bisnis tambahan<br>Nomor telepon nasional<br>Nomor telepon internasional<br>Website bisnis<br>Rating Google<br>Jumlah rating<br>Jam operasional | Place Details Enterprise | $0,020 | 1.000 hit | 0 hit pada flow Prospect Finder utama; 1 hit jika tab Business Info dibuka |
| Place Details - Full Prospect Finder | Semua field Business Info<br>`priceLevel`<br>`utcOffsetMinutes`<br>`editorialSummary`<br>`photos`<br>`delivery`<br>`dineIn`<br>`takeout`<br>`curbsidePickup`<br>`parkingOptions`<br>`paymentOptions`<br>`accessibilityOptions` | Detail lengkap Prospect Finder<br>Fasilitas dan atribut bisnis<br>Metadata foto<br>Jam operasional dan kontak | Place Details Enterprise + Atmosphere | $0,025 | 1.000 hit | 1 hit saat user membuka detail tempat di Prospect Finder |
| Place Photo (New) | Foto binary berdasarkan `photoName` | Menampilkan foto tempat | Place Details Photos | $0,007 | 1.000 hit | 1 hit jika foto pertama tersedia; tambahan 1 hit per foto |
| Custom Search JSON API - Image Search | `title`<br>`link`<br>metadata hasil | Pencarian gambar menu/produk | Custom Search JSON API | $0,005 | 100 query/hari | 0 hit secara default; 1 hit jika user menekan “Load menu photos” |

### Estimasi Flow Standar

| Jenis flow | Hit yang terjadi | Total hit | Estimasi biaya sebelum free tier |
|---|---|---:|---:|
| Search keyword → detail → 1 foto | 1 Text Search + 1 Full Prospect Finder + 1 Place Photo | 3 | $0,064 |
| Search nearby → detail → 1 foto | 1 Nearby Search + 1 Full Prospect Finder + 1 Place Photo | 3 | $0,064 |
| Search keyword → detail tanpa foto | 1 Text Search + 1 Full Prospect Finder | 2 | $0,057 |
| Search keyword → detail → 1 foto → menu image search | 1 Text Search + 1 Full Prospect Finder + 1 Place Photo + 1 Custom Search Image | 4 | $0,069 |
| Flow yang seluruhnya mendapat cache hit | Tidak ada request provider baru | 0 | $0 |

> Rata-rata di atas adalah estimasi per satu kali flow, bukan rata-rata historis dari database. Nilainya memakai asumsi satu search, satu tempat dibuka, satu foto ditampilkan, dan cache MISS.

## Ranking Berdasarkan Harga dan Frekuensi Hit

Tabel berikut diurutkan dari harga per hit paling mahal ke paling murah. Kolom frekuensi menunjukkan penggunaan dalam satu flow standar dengan cache MISS.

| Ranking | API / field group | Harga per hit | Frekuensi tipikal per flow | Dampak biaya dalam flow |
|---:|---|---:|---:|---:|
| 1 | Nearby Search (New): `id`, `displayName`, `formattedAddress`, `primaryTypeDisplayName`, `types`, `businessStatus`, `location` | $0,032 | 1 hit jika user memakai nearby search | $0,032 |
| 1 | Text Search (New): `id`, `displayName`, `formattedAddress`, `primaryTypeDisplayName`, `types`, `businessStatus`, `location` | $0,032 | 1 hit jika user memakai keyword search | $0,032 |
| 2 | Place Details - Full Prospect Finder: kontak, rating, jam buka, foto, fasilitas, dan atribut bisnis | $0,025 | 1 hit saat detail dibuka | $0,025 |
| 3 | Place Details - Business Info: nomor telepon, website, rating, jumlah rating, jam operasional | $0,020 | 0-1 hit; hanya saat tab Business Info digunakan | $0-$0,020 |
| 4 | Place Details - Core: field dasar, `googleMapsUri`, dan metadata `photos` | $0,017 | 0-1 hit; digunakan pada flow Core Detail | $0-$0,017 |
| 5 | Place Photo (New): binary image berdasarkan `photoName` | $0,007 | 1 hit untuk foto pertama; +1 per foto tambahan | $0,007 atau lebih |
| 6 | Custom Search Image: `title`, `link`, metadata hasil gambar | $0,005 | 0 secara default; 1 hit jika “Load menu photos” dipakai | $0-$0,005 |

### Prioritas Biaya dalam Flow Standar

Untuk flow utama `search → buka detail → tampilkan foto pertama`, komposisi biaya biasanya:

| Urutan kontribusi biaya | Komponen | Biaya |
|---:|---|---:|
| 1 | Search: Nearby atau Text Search | $0,032 |
| 2 | Full Prospect Finder Detail | $0,025 |
| 3 | Foto pertama | $0,007 |
|  | **Total satu flow** | **$0,064** |

Dengan demikian, API yang paling mahal per hit sekaligus paling rutin digunakan adalah **Nearby Search/Text Search**, diikuti **Full Prospect Finder Detail**. Komponen yang paling mudah menambah jumlah hit adalah **Place Photo**, karena setiap foto tambahan menambah satu request.

## Field Berdasarkan API

### Nearby Search (New)

Endpoint: `POST https://places.googleapis.com/v1/places:searchNearby`

Field mask yang digunakan:

```text
places.id,
places.displayName,
places.formattedAddress,
places.primaryTypeDisplayName,
places.types,
places.businessStatus,
places.location
```

| Field | Kegunaan |
|---|---|
| `id` | Google Place ID |
| `displayName` | Nama bisnis/tempat |
| `formattedAddress` | Alamat tampilan |
| `primaryTypeDisplayName` | Kategori utama |
| `types` | Tipe/kategori Google |
| `businessStatus` | Status bisnis |
| `location` | Latitude dan longitude |

### Text Search (New)

Endpoint: `POST https://places.googleapis.com/v1/places:searchText`

Field yang diminta sama dengan Nearby Search:

| Field | Kegunaan |
|---|---|
| `id` | Google Place ID |
| `displayName` | Nama bisnis/tempat |
| `formattedAddress` | Alamat tampilan |
| `primaryTypeDisplayName` | Kategori utama |
| `types` | Tipe/kategori Google |
| `businessStatus` | Status bisnis |
| `location` | Latitude dan longitude |

### Place Details - Core

Endpoint: `GET https://places.googleapis.com/v1/places/{placeId}`

Field mask:

```text
id,
displayName,
formattedAddress,
primaryTypeDisplayName,
types,
businessStatus,
googleMapsUri,
location,
photos
```

| Field | Kegunaan |
|---|---|
| `id` | Google Place ID |
| `displayName` | Nama tempat |
| `formattedAddress` | Alamat |
| `primaryTypeDisplayName` | Kategori utama |
| `types` | Tipe/kategori Google |
| `businessStatus` | Status bisnis |
| `googleMapsUri` | Link ke Google Maps |
| `location` | Latitude dan longitude |
| `photos` | Metadata foto yang tersedia |

### Place Details - Business Info

Endpoint: `GET https://places.googleapis.com/v1/places/{placeId}`

Field mask:

```text
id,
displayName,
formattedAddress,
primaryTypeDisplayName,
types,
businessStatus,
googleMapsUri,
location,
nationalPhoneNumber,
internationalPhoneNumber,
websiteUri,
rating,
userRatingCount,
regularOpeningHours
```

| Field | Kegunaan |
|---|---|
| `id` | Google Place ID |
| `displayName` | Nama bisnis |
| `formattedAddress` | Alamat |
| `primaryTypeDisplayName` | Kategori utama |
| `types` | Tipe/kategori Google |
| `businessStatus` | Status bisnis |
| `googleMapsUri` | Link ke Google Maps |
| `location` | Latitude dan longitude |
| `nationalPhoneNumber` | Nomor telepon nasional |
| `internationalPhoneNumber` | Nomor telepon internasional |
| `websiteUri` | Website bisnis |
| `rating` | Rating Google |
| `userRatingCount` | Jumlah rating |
| `regularOpeningHours` | Jam operasional reguler |

### Place Photo (New)

Endpoint:

```text
GET https://places.googleapis.com/v1/{photoName}/media?maxWidthPx=800
```

API ini mengambil binary image berdasarkan resource name dari field `photos`. Aplikasi menggunakan ukuran maksimum `800px`.

### Place Details - Full Prospect Finder

Prospect Finder menggunakan field mask yang lebih luas saat membuka detail langsung:

```text
id, displayName, formattedAddress, primaryTypeDisplayName, types,
businessStatus, rating, userRatingCount, nationalPhoneNumber,
internationalPhoneNumber, websiteUri, googleMapsUri, location,
priceLevel, utcOffsetMinutes, editorialSummary, photos,
regularOpeningHours, delivery, dineIn, takeout, curbsidePickup,
parkingOptions, paymentOptions, accessibilityOptions
```

Field tambahan seperti `delivery`, `dineIn`, `takeout`, `curbsidePickup`, `parkingOptions`, `paymentOptions`, dan `accessibilityOptions` menyebabkan request masuk ke SKU **Place Details Enterprise + Atmosphere** dengan harga **$25 per 1.000 request** atau sekitar **$0,025 per hit**. Free tier SKU ini adalah **1.000 hit per bulan**.

### Custom Search JSON API - Image Search

Endpoint:

```text
GET https://www.googleapis.com/customsearch/v1
```

Parameter utama yang digunakan:

| Parameter | Kegunaan |
|---|---|
| `key` | API key |
| `cx` | Programmable Search Engine ID |
| `searchType=image` | Mengaktifkan pencarian gambar |
| `num` | Jumlah hasil yang diminta |
| `q` | Query pencarian gambar |

Field hasil yang digunakan aplikasi:

| Field | Kegunaan |
|---|---|
| `title` | Judul hasil gambar |
| `link` | URL gambar |

## Contoh Perhitungan Biaya

| Penggunaan | Gross cost sebelum free tier |
|---|---:|
| 1.000 Nearby Search | $32 |
| 1.000 Text Search | $32 |
| 1.000 Place Details Core | $17 |
| 1.000 Place Details Business Info | $20 |
| 1.000 Place Photo | $7 |
| 1.000 Custom Search Image | $5 |

Free tier dikurangi berdasarkan SKU masing-masing. Contoh, 5.000 Nearby Search gratis tidak mengurangi free tier Text Search atau Place Details.

## Hit Aktual per Satu Alur Penggunaan

Tabel berikut menunjukkan jumlah request ke provider ketika cache dalam keadaan **MISS**. Jika cache **HIT**, request ke provider menjadi 0 untuk operasi yang sama.

| Case pengguna | Request provider yang terjadi | Jumlah hit | Estimasi biaya sebelum free tier |
|---|---|---:|---:|
| Search berdasarkan keyword | Text Search Pro | 1 | $0,032 |
| Search berdasarkan lokasi + kategori | Nearby Search Pro | 1 | $0,032 |
| Search yang sama dan masih ada di cache | Tidak ada request provider | 0 | $0 |
| Membuka detail tempat dari Prospect Finder | Place Details Enterprise + Atmosphere (field lengkap) | 1 | $0,025 |
| Membuka detail tempat dengan cache detail masih tersedia | Tidak ada request provider | 0 | $0 |
| Membuka detail Core dari halaman Prospect/Customer | Place Details Pro | 1 | $0,017 |
| Membuka Business Info dari halaman Prospect/Customer | Place Details Enterprise | 1 | $0,020 |
| Detail + otomatis menampilkan 1 foto pertama | Place Details Enterprise + Atmosphere + Place Photo | 1 + 1 | $0,032 atau $0,025 + $0,007 |
| Detail + menampilkan 3 foto | Place Details Enterprise + Atmosphere + 3 Place Photo | 1 + 3 | $0,046 atau $0,025 + (3 x $0,007) |
| Klik “Load one more photo” | Place Photo | 1 per foto | $0,007 per foto |
| Klik “Load menu photos” | Custom Search JSON API Image Search | 1 | $0,005 |
| Search menu yang sama dengan cache provider tersedia | Tidak ada request provider | 0 | $0 |

### Case 1 - Search Keyword

Contoh: Administrator memasukkan keyword `restaurant Jakarta` lalu menekan Search.

```text
1 x Text Search Pro = 1 hit
1 x $0,032 = $0,032
```

Pencarian Text Search saat ini tidak otomatis mengambil halaman berikutnya dari `nextPageToken`, sehingga satu aksi search normalnya hanya menghasilkan satu hit provider.

### Case 2 - Search Nearby

Contoh: Administrator memilih kategori, koordinat, dan radius lalu menekan Search tanpa keyword.

```text
1 x Nearby Search Pro = 1 hit
1 x $0,032 = $0,032
```

Implementasi default menggunakan satu tile/radius pada pencarian interaktif. Jika workflow diubah untuk melakukan tiling atau category fan-out, setiap request tambahan akan menambah satu hit dan biaya yang sama.

### Case 3 - Membuka Detail dari Prospect Finder

Prospect Finder memanggil `DetailFull`, yang merupakan satu request Place Details dengan field Business Info, foto, dan atribut tambahan.

```text
1 x Place Details Enterprise + Atmosphere = 1 hit
1 x $0,025 = $0,025
```

Saat detail terbuka, frontend otomatis mencoba mengambil satu foto pertama jika tersedia:

```text
1 x Place Photo = 1 hit
1 x $0,007 = $0,007
Total = $0,032
```

Jika tidak ada foto, request Place Photo tidak terjadi.

### Case 4 - Membuka Detail Core dan Business Info Terpisah

Pada halaman yang memisahkan data dasar dan Business Info, request terjadi sesuai tab yang dibuka:

| Aksi | Hit | Biaya |
|---|---:|---:|
| Buka Core Detail | 1 Place Details Pro | $0,017 |
| Buka Business Info | 1 Place Details Enterprise | $0,020 |
| Buka keduanya dalam satu sesi | 2 hit | $0,037 |

Jika masing-masing response masih ada di cache, request provider tidak diulang.

### Case 5 - Foto Tambahan

Metadata foto diperoleh dari request Place Details. Media foto baru diambil saat frontend membutuhkan URL foto tersebut.

| Aksi | Hit tambahan | Biaya tambahan |
|---|---:|---:|
| Menampilkan foto pertama | 1 Place Photo | $0,007 |
| Menampilkan 3 foto | 3 Place Photo | $0,021 |
| Klik “Load one more photo” satu kali | 1 Place Photo | $0,007 |
| Foto yang sama sudah tersedia di browser/session cache | 0 | $0 |

### Case 6 - Pencarian Foto Menu

Jika foto menu dari Google Places tidak tersedia atau user menekan `Load menu photos`, aplikasi memanggil Custom Search JSON API dengan `searchType=image`.

```text
1 x Custom Search Image = 1 query
1 x $0,005 = $0,005
```

API ini bukan Google Maps/Places API. Free tier-nya adalah 100 query per hari untuk customer lama, bukan free tier bulanan Places. Google menyatakan Custom Search JSON API tidak tersedia untuk customer baru dan dijadwalkan dihentikan pada 1 Januari 2027.

## Contoh Total Biaya per Sesi

| Skenario sesi | Perhitungan hit | Total estimasi |
|---|---|---:|
| Search keyword saja | 1 Text Search | $0,032 |
| Search nearby saja | 1 Nearby Search | $0,032 |
| Search keyword + buka detail | 1 Text Search + 1 Place Details Enterprise + Atmosphere | $0,057 |
| Search keyword + detail + 1 foto | 1 Text Search + 1 Place Details Enterprise + Atmosphere + 1 Photo | $0,064 |
| Search keyword + detail + 3 foto | 1 Text Search + 1 Place Details Enterprise + Atmosphere + 3 Photo | $0,078 |
| Search keyword + detail + 1 foto + menu image search | 1 Text Search + 1 Full Details + 1 Photo + 1 Custom Search | $0,069 |
| Core Detail + Business Info + 1 foto | 1 Core + 1 Business Info + 1 Photo | $0,044 |

Angka di atas adalah gross estimate sebelum free tier, volume discount, pajak, dan kurs. Dalam billing nyata, hit gratis akan mengurangi jumlah billable event sesuai SKU masing-masing.

## Cache dan Provider Hit

- Cache hit tidak mengirim request baru ke Google.
- Cache hit tidak menambah biaya provider.
- Search, Core Detail, dan Business Info memiliki cache TTL yang dapat dikonfigurasi melalui environment variable.
- Pencarian Text Search tidak otomatis mengikuti `nextPageToken`, sehingga pagination tambahan tidak terjadi pada request awal.
- Photo baru dipanggil ketika aplikasi benar-benar membutuhkan media foto.

## API Google yang Tidak Digunakan

Tidak ditemukan penggunaan langsung untuk:

- Google Maps JavaScript API
- Geocoding API
- Directions API
- Distance Matrix API
- Routes API
- Autocomplete API

Peta pada frontend menggunakan Leaflet dan OpenStreetMap. Link navigasi ke Google Maps hanya membuka URL Google Maps dan bukan API request berbayar.

## Sumber Resmi

- [Google Maps Platform Pricing](https://developers.google.com/maps/billing-and-pricing/pricing)
- [Google Maps Platform SKU Details](https://developers.google.com/maps/billing-and-pricing/sku-details)
- [Places API - Choose Fields](https://developers.google.com/maps/documentation/places/web-service/choose-fields)
- [Places API - Usage and Billing](https://developers.google.com/maps/documentation/places/web-service/usage-and-billing)
- [Custom Search JSON API Pricing](https://developers.google.com/custom-search/v1/overview)

## Disclaimer

Dokumen ini adalah estimasi teknis berdasarkan field mask dan endpoint yang digunakan aplikasi. Tagihan final mengikuti Google Cloud Billing dan dapat berubah jika Google mengubah harga, free tier, SKU, region pricing, atau volume discount.
