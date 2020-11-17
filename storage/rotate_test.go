package storage

import (
	"sync"
	"testing"
)

func TestFileWrite_Write(t *testing.T) {
	var path = "./log/test.log"
	storage := Opt.WithFile(path).Backups(5).Compress(7).MaxSize(10).SaveTime(30).Done()
	wg := &sync.WaitGroup{}
	for gi := 0; gi < 100; gi++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				if _, err := storage.Write([]byte("test data test data test data test data test data test data \n")); err != nil {
					t.Error(err)
				} else {
					//t.Logf("write len %d",n)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
