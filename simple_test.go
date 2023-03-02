package bigcache

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func Test(t *testing.T) {
	hasher := fnv64a{}

	cache, err := New(context.Background(), Config{
		Shards: 1024,
		// time after which entry can be evicted
		LifeWindow: 600 * time.Minute,
		// rps * lifeWindow, used only in initial memory allocation
		CleanWindow: 600 * time.Second,
		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 10485760,
		// prints information about additional memory allocation
		Verbose: true,
		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 102400,
		// callback fired when the oldest entry is removed because of its
		// expiration time or no space left for the new entry. Default value is nil which
		// means no callback and it prevents from unwrapping the oldest entry.
		OnRemove:           nil,
		MaxEntriesInWindow: 50000,
	})
	if err != nil {
		return
	}

	for i := 0; i < 1200; i++ {
		group := sync.WaitGroup{}
		group.Add(180000)
		for j := 0; j < 180000; j++ {
			j := j
			go func() {
				defer group.Done()
				from := fmt.Sprintf("%04s", strconv.FormatInt(int64(j), 32))
				to := fmt.Sprintf("%04s", strconv.FormatInt(int64(j+180000), 32))
				err = cache.Set(fmt.Sprintf("%s#%s", "transfer", from), []byte(fmt.Sprintf("%06d", rand.Intn(10000))))
				err = cache.Set(fmt.Sprintf("%s#%s", "transfer", to), []byte(fmt.Sprintf("%06d", rand.Intn(10000))))
			}()
			go func() {
				err = cache.Set("CONTRACT_MANAGE#ContractByteCode:transfer", []byte("服务使用sdk-go:2.3.1，启动后可以正常处理请求，当调用添加组织根证书成功后，sdk无法处理后续请求，重启服务后可以正常处理请求。（此问题可复现）\n初步猜测是添加组织根证书后，节点的RPC Server服务会重启，导致sdk的client链接出现问题，无法处理后续请求。"))
			}()
			go func() {
				err = cache.Set("stateDBSavePointKey", []byte("225500125635478"))
			}()
			go func() {
				err = cache.Set("CHAIN_CONFIG#CHAIN_CONFIG", []byte(" 在开发中我们经常使用tag保留历史版本，以防新开发功能bug快速回退到上一个版本。每次功能上线前均使用git flow 管理tag , "))
			}()
		}
		group.Wait()
		println(i)
		group1 := sync.WaitGroup{}
		group1.Add(180000)
		for k := 0; k < 180000; k++ {
			k := k
			go func() {
				group1.Done()
				acc1 := fmt.Sprintf("%s#%04s", "transfer", strconv.FormatInt(int64(k), 32))
				acc2 := fmt.Sprintf("%s#%04s", "transfer", strconv.FormatInt(int64(k+180000), 32))
				_, err1 := cache.Get(acc1)
				_, err2 := cache.Get(acc2)
				if err1 != nil {
					sum64 := hasher.Sum64(acc1)
					cache.Get(acc1)
					println(sum64 & 1023)
				}
				if err2 != nil {
					sum64 := hasher.Sum64(acc2)
					cache.Get(acc2)
					println(sum64 & 1023)
				}
				_, err3 := cache.Get("CONTRACT_MANAGE#ContractByteCode:transfer")
				if err3 != nil {
					sum64 := hasher.Sum64("CONTRACT_MANAGE#ContractByteCode:transfer")
					cache.Get("CONTRACT_MANAGE#ContractByteCode:transfer")
					println(sum64 & 1023)
				}
				_, err4 := cache.Get("stateDBSavePointKey")
				if err4 != nil {
					cache.Get("stateDBSavePointKey")
				}
				_, err5 := cache.Get("CHAIN_CONFIG#CHAIN_CONFIG")
				if err5 != nil {
					cache.Get("CHAIN_CONFIG#CHAIN_CONFIG")
				}
			}()
		}
		group1.Wait()
		println(i)
	}
}
