package repositories

import "github.com/topfreegames/Will.IAM/models"

// Services repository
type Services interface {
	List() ([]models.Service, error)
	Get(string) (*models.Service, error)
	WithPermissionName(string) (*models.Service, error)
	Create(*models.Service) error
	Update(*models.Service) error
	Clone() Services
	setStorage(*Storage)
}

type services struct {
	*withStorage
}

func (ss *services) Clone() Services {
	return NewServices(ss.storage.Clone())
}

func (ss services) Create(s *models.Service) error {
	_, err := ss.storage.PG.DB.Query(
		s, `INSERT INTO services (name, permission_name, service_account_id,
		creator_service_account_id, am_url) VALUES (?name, ?permission_name,
		?service_account_id, ?creator_service_account_id, ?am_url) RETURNING id`,
		s,
	)
	return err
}

// List returns all services in storage
func (ss services) List() ([]models.Service, error) {
	var allServices []models.Service
	if _, err := ss.storage.PG.DB.Query(
		&allServices, `SELECT * FROM services`,
	); err != nil {
		return nil, err
	}
	return allServices, nil
}

// Get service
func (ss services) Get(id string) (*models.Service, error) {
	s := new(models.Service)
	if _, err := ss.storage.PG.DB.Query(
		s, `SELECT * FROM services WHERE id = ?`, id,
	); err != nil {
		return nil, err
	}
	return s, nil
}

// WithPermissionName looks for a service given a PermissionName
func (ss services) WithPermissionName(
	permissionName string,
) (*models.Service, error) {
	s := new(models.Service)
	if _, err := ss.storage.PG.DB.Query(
		s, `SELECT * FROM services WHERE permission_name = ?`, permissionName,
	); err != nil {
		return nil, err
	}
	return s, nil
}

func (ss services) Update(s *models.Service) error {
	_, err := ss.storage.PG.DB.Exec(
		`UPDATE services SET name = ?name, permission_name = ?permission_name,
		am_url = ?am_url WHERE id = ?id`, s,
	)
	return err
}

// NewServices services ctor
func NewServices(s *Storage) Services {
	return &services{&withStorage{storage: s}}
}
