package domains

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/maskrapp/api/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Domains struct {
	db       *gorm.DB
	interval time.Duration
	mutex    sync.RWMutex
	domains  []*models.Domain
}

// New creates a new Domains instance.
func New(db *gorm.DB, interval time.Duration) *Domains {
	return &Domains{
		db:       db,
		interval: interval,
		mutex:    sync.RWMutex{},
		domains:  make([]*models.Domain, 0),
	}
}

func (d *Domains) update() {
	var domains []*models.Domain
	err := d.db.Find(&domains).Error
	if err != nil {
		logrus.Errorf("db error(updateAvailableDomains): %v", err)
		return
	}
	d.mutex.Lock()
	d.domains = domains
	d.mutex.Unlock()
	logrus.Debugf("available domains: %v", d.domains)
}

// Get retrieves a domain under the given input.
func (d *Domains) Get(domain string) (*models.Domain, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, v := range d.domains {
		if strings.EqualFold(v.Domain, domain) {
			return v, nil
		}
	}
	return nil, errors.New("domain not found")
}

// Values returns the recently fetched domains.
func (d *Domains) Values() []*models.Domain {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.domains
}

// Start starts the domain fetching task.
func (d *Domains) Start() {
	go func() {
		for {
			d.update()
			time.Sleep(d.interval)
		}
	}()
}
