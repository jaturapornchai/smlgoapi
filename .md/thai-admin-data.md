# 🌏 Thai Administrative Data API Documentation

## Overview

API endpoints for retrieving Thai administrative data including provinces, districts (amphures), sub-districts (tambons), and postal code lookup.

## Base URL

`http://localhost:8008/v1`

---

## 📋 Available Endpoints

### 1. POST `/provinces`

Get all Thai provinces.

#### Request Format

```json
{}
```

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/provinces" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```javascript
const getProvinces = async () => {
  const response = await fetch("http://localhost:8008/v1/provinces", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  return await response.json();
};
```

#### Response Format

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name_th": "กรุงเทพมหานคร",
      "name_en": "Bangkok",
      "geography_id": 2,
      "created_at": "2019-08-09T03:33:09Z",
      "updated_at": "2022-05-16T06:31:26Z",
      "deleted_at": null
    },
    {
      "id": 2,
      "name_th": "สมุทรปราการ",
      "name_en": "Samut Prakan",
      "geography_id": 2,
      "created_at": "2019-08-09T03:33:09Z",
      "updated_at": "2022-05-16T06:31:26Z",
      "deleted_at": null
    }
  ]
}
```

---

### 2. POST `/amphures`

Get districts (amphures) in a specific province.

#### Request Format

```json
{
  "province_id": number
}
```

#### Parameters

| Parameter     | Type   | Required | Description |
| ------------- | ------ | -------- | ----------- |
| `province_id` | number | ✅ Yes   | Province ID |

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/amphures" \
  -H "Content-Type: application/json" \
  -d '{
    "province_id": 1
  }'
```

```python
def get_amphures(province_id):
    url = "http://localhost:8008/v1/amphures"
    payload = {"province_id": province_id}
    response = requests.post(url, json=payload)
    return response.json()

# Usage - Get Bangkok districts
bangkok_districts = get_amphures(1)
```

#### Response Format

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name_th": "พระนคร",
      "name_en": "Phra Nakhon",
      "province_id": 1,
      "created_at": "2019-08-09T03:33:09Z",
      "updated_at": "2022-05-16T06:31:26Z",
      "deleted_at": null
    },
    {
      "id": 2,
      "name_th": "ดุสิต",
      "name_en": "Dusit",
      "province_id": 1,
      "created_at": "2019-08-09T03:33:09Z",
      "updated_at": "2022-05-16T06:31:26Z",
      "deleted_at": null
    }
  ]
}
```

---

### 3. POST `/tambons`

Get sub-districts (tambons) in a specific amphure.

#### Request Format

```json
{
  "province_id": number,
  "amphure_id": number
}
```

#### Parameters

| Parameter     | Type   | Required | Description           |
| ------------- | ------ | -------- | --------------------- |
| `province_id` | number | ✅ Yes   | Province ID           |
| `amphure_id`  | number | ✅ Yes   | Amphure (district) ID |

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/tambons" \
  -H "Content-Type: application/json" \
  -d '{
    "province_id": 1,
    "amphure_id": 1
  }'
```

```javascript
const getTambons = async (provinceId, amphureId) => {
  const response = await fetch("http://localhost:8008/v1/tambons", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      province_id: provinceId,
      amphure_id: amphureId,
    }),
  });
  return await response.json();
};

// Usage - Get sub-districts in Phra Nakhon, Bangkok
const tambons = await getTambons(1, 1);
```

#### Response Format

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": 10200,
      "name_th": "พระบรมมหาราชวัง",
      "name_en": "Phra Borom Maha Ratchawang",
      "amphure_id": 1,
      "created_at": "2019-08-09T03:33:09Z",
      "updated_at": "2022-05-16T06:31:26Z",
      "deleted_at": null
    },
    {
      "id": 2,
      "zip_code": 10200,
      "name_th": "วังบูรพาภิรมย์",
      "name_en": "Wang Burapha Phirom",
      "amphure_id": 1,
      "created_at": "2019-08-09T03:33:09Z",
      "updated_at": "2022-05-16T06:31:26Z",
      "deleted_at": null
    }
  ]
}
```

---

### 4. POST `/findbyzipcode`

Find location information by postal code.

#### Request Format

```json
{
  "zip_code": "string"
}
```

#### Parameters

| Parameter  | Type   | Required | Description                 |
| ---------- | ------ | -------- | --------------------------- |
| `zip_code` | string | ✅ Yes   | Thai postal code (5 digits) |

#### Usage Examples

```bash
curl -X POST "http://localhost:8008/v1/findbyzipcode" \
  -H "Content-Type: application/json" \
  -d '{
    "zip_code": "10110"
  }'
```

```python
def find_by_zipcode(zip_code):
    url = "http://localhost:8008/v1/findbyzipcode"
    payload = {"zip_code": zip_code}
    response = requests.post(url, json=payload)
    return response.json()

# Usage
location = find_by_zipcode("10110")
print(f"Province: {location['data'][0]['province_name_en']}")
```

#### Response Format

```json
{
  "success": true,
  "data": [
    {
      "province_name_th": "กรุงเทพมหานคร",
      "province_name_en": "Bangkok",
      "amphure_name_th": "พระนคร",
      "amphure_name_en": "Phra Nakhon",
      "tambon_name_th": "พระบรมมหาราชวัง",
      "tambon_name_en": "Phra Borom Maha Ratchawang",
      "zip_code": 10200
    }
  ]
}
```

---

## 🔧 Integration Examples

### Complete Address Lookup System

```javascript
class ThaiAddressAPI {
  constructor(baseUrl = "http://localhost:8008/v1") {
    this.baseUrl = baseUrl;
  }

  async getProvinces() {
    const response = await fetch(`${this.baseUrl}/provinces`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
    return await response.json();
  }

  async getAmphures(provinceId) {
    const response = await fetch(`${this.baseUrl}/amphures`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ province_id: provinceId }),
    });
    return await response.json();
  }

  async getTambons(provinceId, amphureId) {
    const response = await fetch(`${this.baseUrl}/tambons`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        province_id: provinceId,
        amphure_id: amphureId,
      }),
    });
    return await response.json();
  }

  async findByZipcode(zipCode) {
    const response = await fetch(`${this.baseUrl}/findbyzipcode`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ zip_code: zipCode }),
    });
    return await response.json();
  }

  // Helper method to get complete hierarchy
  async getCompleteAddress(zipCode) {
    const location = await this.findByZipcode(zipCode);

    if (location.success && location.data.length > 0) {
      const addr = location.data[0];
      return {
        province: {
          th: addr.province_name_th,
          en: addr.province_name_en,
        },
        amphure: {
          th: addr.amphure_name_th,
          en: addr.amphure_name_en,
        },
        tambon: {
          th: addr.tambon_name_th,
          en: addr.tambon_name_en,
        },
        zipcode: addr.zip_code,
      };
    }
    return null;
  }
}

// Usage
const addressAPI = new ThaiAddressAPI();

// Get all provinces
const provinces = await addressAPI.getProvinces();

// Get districts in Bangkok
const bangkokDistricts = await addressAPI.getAmphures(1);

// Get complete address by zipcode
const address = await addressAPI.getCompleteAddress("10110");
console.log(address);
```

### PHP Address Form Helper

```php
<?php
class ThaiAddressHelper {
    private $baseUrl;

    public function __construct($baseUrl = "http://localhost:8008/v1") {
        $this->baseUrl = $baseUrl;
    }

    private function makeRequest($endpoint, $data = []) {
        $url = $this->baseUrl . $endpoint;
        $ch = curl_init();

        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
        curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

        $response = curl_exec($ch);
        curl_close($ch);

        return json_decode($response, true);
    }

    public function getProvinceOptions() {
        $result = $this->makeRequest('/provinces', []);
        $options = '<option value="">เลือกจังหวัด</option>';

        if ($result['success']) {
            foreach ($result['data'] as $province) {
                $options .= sprintf(
                    '<option value="%d">%s</option>',
                    $province['id'],
                    $province['name_th']
                );
            }
        }

        return $options;
    }

    public function getAmphureOptions($provinceId) {
        $result = $this->makeRequest('/amphures', ['province_id' => $provinceId]);
        $options = '<option value="">เลือกอำเภอ</option>';

        if ($result['success']) {
            foreach ($result['data'] as $amphure) {
                $options .= sprintf(
                    '<option value="%d">%s</option>',
                    $amphure['id'],
                    $amphure['name_th']
                );
            }
        }

        return $options;
    }

    public function autoFillByZipcode($zipCode) {
        $result = $this->makeRequest('/findbyzipcode', ['zip_code' => $zipCode]);

        if ($result['success'] && !empty($result['data'])) {
            return $result['data'][0];
        }

        return null;
    }
}

// Usage in form
$addressHelper = new ThaiAddressHelper();
echo $addressHelper->getProvinceOptions();
?>
```

### Python Address Validator

```python
class ThaiAddressValidator:
    def __init__(self, base_url="http://localhost:8008/v1"):
        self.base_url = base_url

    def validate_zipcode(self, zip_code):
        """Validate Thai postal code"""
        if not zip_code or len(zip_code) != 5 or not zip_code.isdigit():
            return False, "Invalid zipcode format"

        try:
            response = requests.post(
                f"{self.base_url}/findbyzipcode",
                json={"zip_code": zip_code}
            )
            result = response.json()

            if result['success'] and result['data']:
                return True, result['data'][0]
            else:
                return False, "Zipcode not found"
        except Exception as e:
            return False, f"API error: {str(e)}"

    def validate_address_hierarchy(self, province_id, amphure_id, tambon_id):
        """Validate address hierarchy consistency"""
        try:
            # Check if amphure belongs to province
            amphures = requests.post(
                f"{self.base_url}/amphures",
                json={"province_id": province_id}
            ).json()

            valid_amphure_ids = [a['id'] for a in amphures['data']]
            if amphure_id not in valid_amphure_ids:
                return False, "Amphure does not belong to the specified province"

            # Check if tambon belongs to amphure
            tambons = requests.post(
                f"{self.base_url}/tambons",
                json={"province_id": province_id, "amphure_id": amphure_id}
            ).json()

            valid_tambon_ids = [t['id'] for t in tambons['data']]
            if tambon_id not in valid_tambon_ids:
                return False, "Tambon does not belong to the specified amphure"

            return True, "Address hierarchy is valid"
        except Exception as e:
            return False, f"Validation error: {str(e)}"

# Usage
validator = ThaiAddressValidator()

# Validate zipcode
is_valid, result = validator.validate_zipcode("10110")
if is_valid:
    print(f"Valid zipcode: {result['province_name_en']}")

# Validate address hierarchy
is_valid, message = validator.validate_address_hierarchy(1, 1, 1)
print(message)
```

---

## 🔍 Use Cases

### E-commerce Address Forms

- Dynamic dropdowns for province → amphure → tambon
- Auto-fill address fields using postal code
- Address validation before order submission

### Shipping & Logistics

- Calculate shipping zones based on provinces
- Validate delivery addresses
- Generate shipping labels with complete addresses

### Government Services

- Citizen registration systems
- Tax collection by administrative regions
- Statistical reporting by geographic areas

### Business Intelligence

- Sales analysis by provinces/regions
- Customer demographics mapping
- Market penetration analysis

---

## 📊 Data Coverage

### Complete Dataset

- **Provinces:** 77 provinces
- **Amphures:** 928 districts
- **Tambons:** 7,436 sub-districts
- **Postal Codes:** Full coverage of Thai postal system

### Data Accuracy

- Official government data sources
- Regular updates and maintenance
- Both Thai and English names provided
- Hierarchical relationships maintained

---

## 🚨 Error Handling

### Common Error Responses

```json
{
  "success": false,
  "message": "Province not found"
}
```

```json
{
  "success": false,
  "message": "Invalid zipcode format"
}
```

### Error Types

| Error               | Description                        | Solution                        |
| ------------------- | ---------------------------------- | ------------------------------- |
| Invalid province_id | Province ID doesn't exist          | Check valid province IDs        |
| Invalid amphure_id  | Amphure doesn't belong to province | Verify hierarchy                |
| Zipcode not found   | Postal code doesn't exist          | Check zipcode format            |
| Missing parameters  | Required fields not provided       | Include all required parameters |

---

## 📈 Performance

- **Response Time:** ~50-200ms per request
- **Data Caching:** Optimized for fast lookups
- **Database Indexing:** All geographic relationships indexed
- **Concurrent Requests:** Supports high concurrent usage
