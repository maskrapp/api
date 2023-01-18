package domains

import (
	"errors"
	"sync"
	"time"

	"github.com/maskrapp/common/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Domains struct {
	db           *gorm.DB
	interval     time.Duration
	mutex        sync.RWMutex
	domains      map[string]*models.Domain
	shutdownChan chan struct{}
}

// New creates a new Domains instance.
func New(db *gorm.DB, interval time.Duration) *Domains {
	domains := &Domains{
		db:       db,
		interval: interval,
		mutex:    sync.RWMutex{},
	}
	go domains.startTask()

	return domains
}

// Shutdown stops the domain fetching task.
func (d *Domains) Shutdown() {
	d.shutdownChan <- struct{}{}
}

func (d *Domains) updateDomains() {
	logrus.Debug("Updating domains")
	var domains []*models.Domain
	err := d.db.Find(&domains).Error
	if err != nil {
		logrus.Errorf("db error(updateAvailableDomains): %v", err)
		return
	}
	domainsMap := make(map[string]*models.Domain)
	for _, v := range domains {
		domainsMap[v.Domain] = v
	}
	d.mutex.Lock()
	d.domains = domainsMap
	d.mutex.Unlock()
	logrus.Debugf("available domains: %v", d.domains)
}

// Get retrieves a domain under the given input.
func (d *Domains) Get(domain string) (*models.Domain, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	value, ok := d.domains[domain]
	if !ok {
		return nil, errors.New("domain not found")
	}
	return value, nil
}

// Values returns the recently fetched domains.
func (d *Domains) Values() []*models.Domain {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	values := make([]*models.Domain, 0)
	for _, v := range d.domains {
		values = append(values, v)
	}
	return values
}

// startTask starts the domain fetching task.
func (d *Domains) startTask() {
	go func() {
		for {
			d.updateDomains()
			time.Sleep(d.interval)
		}
	}()
	<-d.shutdownChan
}
