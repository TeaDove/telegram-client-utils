package memory

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	interfaceStorage "github.com/teadove/goteleout/internal/service/storage"

	"github.com/teadove/goteleout/internal/utils"
)

func saveLoad(t *testing.T) *Storage {
	filename := fmt.Sprintf("/tmp/%s.json", uuid.New().String())
	storage := MustNew(true, filename)

	err := storage.Save("key", []byte("value"))
	utils.Check(err)

	res, err := storage.Load("key")
	utils.Check(err)

	assert.Equal(t, []byte("value"), res)

	return storage
}

func TestUnit_MemoryStorage_SaveLoad_Ok(t *testing.T) {
	saveLoad(t)
}

func TestUnit_MemoryStorage_Flush_Ok(t *testing.T) {
	storage := saveLoad(t)

	err := storage.flush()
	utils.Check(err)
}

func TestUnit_MemoryStorage_loadFlushed_Ok(t *testing.T) {
	storage := saveLoad(t)
	err := storage.Save("abc", []byte("def"))
	utils.Check(err)
	oldMapping := storage.mapping

	err = storage.flush()
	utils.Check(err)

	err = storage.loadFlushed()
	utils.Check(err)

	assert.Equal(t, oldMapping, storage.mapping)
}

func TestUnit_MemoryStorage_delete_Ok(t *testing.T) {
	storage := saveLoad(t)

	err := storage.Delete("key")
	utils.Check(err)

	_, err = storage.Load("key")
	assert.Error(t, err, interfaceStorage.KeyError)
}

func saveFlushLoad(wg *sync.WaitGroup, storage *Storage, t *testing.T) {
	defer wg.Done()

	key, value := uuid.New().String(), []byte(uuid.New().String())
	err := storage.Save(key, value)
	utils.Check(err)
	err = storage.flush()
	utils.Check(err)
	err = storage.loadFlushed()
	utils.Check(err)

	storedValue, err := storage.Load(key)
	utils.Check(err)
	assert.Equal(t, value, storedValue)
}

func TestUnit_MemoryStorage_multipleSave_Ok(t *testing.T) {
	storage := saveLoad(t)

	wg := sync.WaitGroup{}
	for i := 0; i < 1_000; i++ {
		wg.Add(1)
		saveFlushLoad(&wg, storage, t)
	}
	wg.Wait()

	err := storage.loadFlushed()
	utils.Check(err)

	assert.Equal(t, 1_001, len(storage.mapping))
}
