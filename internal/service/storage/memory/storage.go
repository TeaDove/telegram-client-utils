package memory

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/utils"
)

type Storage struct {
	mapping   map[string][]byte
	mappingMu *sync.RWMutex
	flushMu   *sync.Mutex

	persistent bool
	needFlush  bool
	filename   string
}

func MustNew(persistent bool, filename string) *Storage {
	mapping := make(map[string][]byte, 30)
	memoryStorage := Storage{mapping, &sync.RWMutex{}, &sync.Mutex{}, persistent, false, filename}

	scheduler := gocron.NewScheduler(time.UTC)
	if memoryStorage.persistent {
		err := memoryStorage.loadFlushed()
		utils.Check(err)
		_, err = scheduler.Every(1 * time.Minute).
			StartAt(time.Now().Add(1 * time.Minute)).
			Do(memoryStorage.flush)
		utils.Check(err)
		scheduler.StartAsync()
		log.Info().Str("status", "flush.scheduled").Str("filename", filename).Send()
	}

	return &memoryStorage
}

func (r *Storage) loadFlushed() error {
	r.flushMu.Lock()
	defer r.flushMu.Unlock()

	content, err := os.ReadFile(r.filename)
	if err != nil {
		log.Warn().Str("status", "err.while.reading.file").Stack().Err(err).Send()
		return nil
	}
	newMap := make(map[string][]byte, 10)
	err = json.Unmarshal(content, &newMap)
	if err != nil {
		log.Warn().
			Str("status", "err.while.unmarshalling.json.recreating.file").
			Stack().
			Err(err).
			Send()
		err = os.Remove(r.filename)
		return err
	}

	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()
	r.mapping = newMap
	log.Info().Str("status", "load_flushed.end").Int("len", len(r.mapping)).Send()
	return err
}

func (r *Storage) flush() error {
	if !r.needFlush {
		log.Info().Str("status", "flush.no.need").Send()
		return nil
	}
	log.Info().Str("status", "flush.begin").Send()

	defer func() { r.needFlush = false }()
	r.flushMu.Lock()
	defer r.flushMu.Unlock()

	jsonMap, err := json.Marshal(r.mapping)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(r.filename, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	_, err = f.Write(jsonMap)
	if err != nil {
		return err
	}

	log.Info().Str("status", "flush.end").Int("len", len(r.mapping)).Send()

	return nil
}

func (r *Storage) Load(k string) ([]byte, error) {
	r.mappingMu.RLock()
	defer r.mappingMu.RUnlock()

	v, ok := r.mapping[k]
	if !ok {
		return []byte{}, storage.KeyError
	}
	return v, nil
}

func (r *Storage) Save(k string, v []byte) error {
	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()

	r.needFlush = true
	r.mapping[k] = v
	return nil
}

func (r *Storage) Delete(k string) error {
	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()

	r.needFlush = true
	delete(r.mapping, k)
	return nil
}
