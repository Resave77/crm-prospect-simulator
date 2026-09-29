# Prospect UI Update Audit

Tanggal audit: 2026-09-29  
Branch sumber: `feature/master-data-ui-sync`  
Branch rilis perubahan: `feature/prospect-ui-sync-2026-09-29`

## File dalam pembaruan

- `frontend/src/views/Sales/Prospect/SalesPipelineView.vue`
- `frontend/src/views/Sales/Dashboard/SalesDashboardView.vue`
- `frontend/src/views/Admin/Prospect/ProspectReviewView.vue`
- `frontend/src/views/Admin/Prospect/ProspectPipelineView.vue`
- `frontend/src/layouts/AdminLayout.vue`
- `frontend/src/components/sales/pipeline/NewLeadGroupCard.vue`
- `frontend/src/views/Admin/Prospect/ProspectFinderView.vue`
- `frontend/tests/prospectFinderMultiResult.test.mjs`

## Ringkasan perubahan

- Sinkronisasi tampilan dan interaksi prospect pipeline untuk Admin dan Sales.
- Penyederhanaan alur tampilan pipeline prospect dan navigasi kartu prospect.
- Penyesuaian review prospect, dashboard sales, dan layout admin untuk navigasi mobile.
- Menjaga dukungan detail prospect, assignment sales, status pipeline, dan pencarian prospect.
- Mempertahankan test Prospect Finder multi-result sebagai pemeriksaan regresi.

## Panduan pull untuk developer lain

```bash
git fetch origin
git switch feature/prospect-ui-sync-2026-09-29
git pull --ff-only origin feature/prospect-ui-sync-2026-09-29
```

Jika sedang mengerjakan perubahan lokal, simpan pekerjaan terlebih dahulu dengan commit atau stash. Jangan melakukan merge manual ke branch kerja sebelum membandingkan delapan file di atas.

## Pemeriksaan konflik

- Perubahan di luar delapan file di atas tidak dimasukkan ke commit pembaruan ini.
- File `backend/.tmp-go-cache/` dan file lokal lain yang belum terkait tidak ikut dipush.
- Developer yang mengambil branch ini sebaiknya menjalankan test frontend sebelum menggabungkannya ke branch utama.
