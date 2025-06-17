# Thai Province Data JSON Files

This folder contains comprehensive Thai geographical data in JSON format, downloaded from the [thai-province-data](https://github.com/kongvut/thai-province-data) repository.

## Files Description

### 📋 `api_province.json` (14.9 KB)
- **Content**: 77 Thai provinces
- **Fields**: `id`, `name_th`, `name_en`, `geography_id`, `created_at`, `updated_at`, `deleted_at`
- **Use Case**: Province lookup, basic administrative data

### 🏘️ `api_amphure.json` (183.6 KB)
- **Content**: All amphures (districts) in Thailand
- **Fields**: `id`, `name_th`, `name_en`, `province_id`, `created_at`, `updated_at`, `deleted_at`
- **Use Case**: District-level administrative data

### 🏡 `api_tambon.json` (1.6 MB)
- **Content**: All tambons (sub-districts) in Thailand
- **Fields**: `id`, `name_th`, `name_en`, `amphure_id`, `created_at`, `updated_at`, `deleted_at`
- **Use Case**: Sub-district level administrative data

### 🌐 `api_province_with_amphure_tambon.json` (1.8 MB)
- **Content**: Complete hierarchical data (Province → Amphure → Tambon)
- **Structure**: Nested JSON with full geographical hierarchy
- **Use Case**: Complete geographical lookups, address validation

### 🔄 `api_revert_tambon_with_amphure_province.json` (4.7 MB)
- **Content**: Reverse lookup from tambon to province
- **Structure**: Each tambon with its parent amphure and province data
- **Use Case**: Address resolution from smallest to largest administrative unit

## Data Structure Examples

### Province Structure
```json
{
  "id": 1,
  "name_th": "กรุงเทพมหานคร",
  "name_en": "Bangkok",
  "geography_id": 2,
  "created_at": "2019-08-09T03:33:09.000+07:00",
  "updated_at": "2022-05-16T06:31:03.000+07:00",
  "deleted_at": null
}
```

### Geography IDs
- `1`: ภาคเหนือ (Northern Thailand)
- `2`: ภาคกลาง (Central Thailand)
- `3`: ภาคตะวันออกเฉียงเหนือ (Northeastern Thailand)
- `4`: ภาคตะวันตก (Western Thailand)
- `5`: ภาคตะวันออก (Eastern Thailand)
- `6`: ภาคใต้ (Southern Thailand)

## Usage in SMLGOAPI

These files can be used for:
- 🔍 **Search Enhancement**: Improve location-based searches
- 📍 **Address Validation**: Validate Thai addresses in API requests
- 🗺️ **Geographical Queries**: Support province/district-based filtering
- 🌏 **Location Services**: Provide administrative boundary information
- 🏢 **Business Logic**: Support location-aware features

## API Integration Ideas

1. **Province Endpoint**: `/api/provinces` - List all provinces
2. **District Endpoint**: `/api/amphures/{province_id}` - Get districts by province
3. **Address Search**: `/api/search/address?query=...` - Fuzzy address search
4. **Reverse Lookup**: `/api/location/{tambon_id}` - Get full hierarchy

## File Size Summary
- Total: ~8.1 MB
- All 77 Thai provinces
- Complete administrative hierarchy
- Thai and English names
- Ready for production use

---
*Downloaded on: June 17, 2025*  
*Source: https://github.com/kongvut/thai-province-data*
