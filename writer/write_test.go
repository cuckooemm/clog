package writer

import (
	"fmt"
	"sync"
	"testing"
)

func TestFileWrite_Write(t *testing.T) {
	var path = "./log/test.log"
	write := NewWrite(path).WithBackups(5).WithCompress().WithMaxAge(0).WithMaxSize(1024 * 512).Done()

	//os.RemoveAll(filepath.Dir(path))
	wg := &sync.WaitGroup{}
	for gi := 0; gi < 100; gi++ {
		wg.Add(1)
		idx := gi
		go func() {
			for i := 0; i < 1000; i++ {
				if _, err := write.Write([]byte(fmt.Sprintf("goroutine [%d] weite idx [%d]\n", idx, i))); err != nil {
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
