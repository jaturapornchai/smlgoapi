package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"smlgoapi/models"
)

// ThaiAdminService handles Thai administrative data operations
type ThaiAdminService struct {
	provincesData          []models.Province
	amphuresData           []models.Amphure
	tambonsData            []models.Tambon
	provincesLoaded        bool
	amphuresLoaded         bool
	tambonsLoaded          bool
	completeLocationData   []models.CompleteLocationData
	completeLocationLoaded bool
}

// NewThaiAdminService creates a new Thai administrative service
func NewThaiAdminService() *ThaiAdminService {
	return &ThaiAdminService{}
}

// loadProvinces loads province data from JSON file
func (s *ThaiAdminService) loadProvinces() error {
	if s.provincesLoaded {
		return nil
	}

	filePath := filepath.Join("provinces", "api_province.json")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read provinces file: %v", err)
	}

	err = json.Unmarshal(data, &s.provincesData)
	if err != nil {
		return fmt.Errorf("failed to parse provinces JSON: %v", err)
	}

	s.provincesLoaded = true
	return nil
}

// loadAmphures loads amphure data from JSON file
func (s *ThaiAdminService) loadAmphures() error {
	if s.amphuresLoaded {
		return nil
	}

	filePath := filepath.Join("provinces", "api_amphure.json")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read amphures file: %v", err)
	}

	err = json.Unmarshal(data, &s.amphuresData)
	if err != nil {
		return fmt.Errorf("failed to parse amphures JSON: %v", err)
	}

	s.amphuresLoaded = true
	return nil
}

// loadTambons loads tambon data from JSON file
func (s *ThaiAdminService) loadTambons() error {
	if s.tambonsLoaded {
		return nil
	}

	filePath := filepath.Join("provinces", "api_tambon.json")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read tambons file: %v", err)
	}

	err = json.Unmarshal(data, &s.tambonsData)
	if err != nil {
		return fmt.Errorf("failed to parse tambons JSON: %v", err)
	}

	s.tambonsLoaded = true
	return nil
}

// GetProvinces returns all provinces
func (s *ThaiAdminService) GetProvinces() ([]models.Province, error) {
	err := s.loadProvinces()
	if err != nil {
		return nil, err
	}

	// Return only essential fields as specified in the docs
	var result []models.Province
	for _, province := range s.provincesData {
		result = append(result, models.Province{
			ID:     province.ID,
			NameTh: province.NameTh,
			NameEn: province.NameEn,
		})
	}

	return result, nil
}

// GetAmphuresByProvinceID returns all amphures for a given province
func (s *ThaiAdminService) GetAmphuresByProvinceID(provinceID int) ([]models.Amphure, error) {
	err := s.loadAmphures()
	if err != nil {
		return nil, err
	}

	var result []models.Amphure
	for _, amphure := range s.amphuresData {
		if amphure.ProvinceID == provinceID {
			result = append(result, models.Amphure{
				ID:     amphure.ID,
				NameTh: amphure.NameTh,
				NameEn: amphure.NameEn,
			})
		}
	}

	return result, nil
}

// GetTambonsByAmphureAndProvince returns all tambons for a given amphure and province
func (s *ThaiAdminService) GetTambonsByAmphureAndProvince(amphureID, provinceID int) ([]models.Tambon, error) {
	err := s.loadTambons()
	if err != nil {
		return nil, err
	}

	// First verify the amphure belongs to the province
	err = s.loadAmphures()
	if err != nil {
		return nil, err
	}

	var amphureFound bool
	for _, amphure := range s.amphuresData {
		if amphure.ID == amphureID && amphure.ProvinceID == provinceID {
			amphureFound = true
			break
		}
	}

	if !amphureFound {
		return nil, fmt.Errorf("amphure_id %d not found in province_id %d", amphureID, provinceID)
	}

	var result []models.Tambon
	for _, tambon := range s.tambonsData {
		if tambon.AmphureID == amphureID {
			result = append(result, models.Tambon{
				ID:     tambon.ID,
				NameTh: tambon.NameTh,
				NameEn: tambon.NameEn,
			})
		}
	}

	return result, nil
}

// loadCompleteLocationData loads complete location data from JSON file
func (s *ThaiAdminService) loadCompleteLocationData() error {
	if s.completeLocationLoaded {
		return nil
	}

	filePath := filepath.Join("provinces", "api_revert_tambon_with_amphure_province.json")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read complete location file: %v", err)
	}

	var tambonsWithNested []models.TambonWithNested
	err = json.Unmarshal(data, &tambonsWithNested)
	if err != nil {
		return fmt.Errorf("failed to parse complete location JSON: %v", err)
	}

	// Convert to our CompleteLocationData structure
	s.completeLocationData = make([]models.CompleteLocationData, len(tambonsWithNested))
	for i, tambon := range tambonsWithNested {
		s.completeLocationData[i] = models.CompleteLocationData{
			Province: models.Province{
				ID:     tambon.Amphure.Province.ID,
				NameTh: tambon.Amphure.Province.NameTh,
				NameEn: tambon.Amphure.Province.NameEn,
			},
			Amphure: models.Amphure{
				ID:     tambon.Amphure.ID,
				NameTh: tambon.Amphure.NameTh,
				NameEn: tambon.Amphure.NameEn,
			},
			Tambon: models.Tambon{
				ID:      tambon.ID,
				NameTh:  tambon.NameTh,
				NameEn:  tambon.NameEn,
				ZipCode: tambon.ZipCode,
			},
		}
	}

	s.completeLocationLoaded = true
	return nil
}

// FindByZipCode finds all locations with the given zip code
func (s *ThaiAdminService) FindByZipCode(zipCode int) ([]models.CompleteLocationData, error) {
	err := s.loadCompleteLocationData()
	if err != nil {
		return nil, err
	}

	var result []models.CompleteLocationData
	for _, location := range s.completeLocationData {
		if location.Tambon.ZipCode == zipCode {
			result = append(result, models.CompleteLocationData{
				Province: models.Province{
					ID:     location.Province.ID,
					NameTh: location.Province.NameTh,
					NameEn: location.Province.NameEn,
				},
				Amphure: models.Amphure{
					ID:     location.Amphure.ID,
					NameTh: location.Amphure.NameTh,
					NameEn: location.Amphure.NameEn,
				},
				Tambon: models.Tambon{
					ID:      location.Tambon.ID,
					NameTh:  location.Tambon.NameTh,
					NameEn:  location.Tambon.NameEn,
					ZipCode: location.Tambon.ZipCode,
				},
			})
		}
	}

	return result, nil
}
