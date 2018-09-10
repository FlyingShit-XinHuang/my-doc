## 数据结构

```
type Map struct {
    // 对dirty的锁
    mu Mutex
    read atomic.Value // readOnly
    dirty map[interface{}]*entry
    // misses记录从Map.read中读取未命中的次数，大于一定值时，将dirty提升为read
    misses int
}

// Map.read的结构
type readOnly struct {
    m       map[interface{}]*entry
    // Store时，entry存入Map.dirty，此时amended为true，表示Map.dirty中存在多余的key
    amended bool
}

type entry struct {
    // p为value的指针
    // 如p为expunged表示已删除，且Map.read中存在但Map.dirty中不存在
    // 如p为nil表示已删除
    p unsafe.Pointer // *interface{}
}
```

## read

```
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
        // 优先从Map.read中读取
        read, _ := m.read.Load().(readOnly)
        e, ok := read.m[key]
        if !ok && read.amended {
                // 不存在且Map.dirty中存在其他keys时，Map.mu加锁
                m.mu.Lock()
                // 二次读取Map.read（因为加锁之前和判断不存在这段时间内，Map.read可能更新）
                read, _ = m.read.Load().(readOnly)
                e, ok = read.m[key]
                if !ok && read.amended {
                        // 仍不存在时，从Map.dirty中读
                        e, ok = m.dirty[key]
                        // Map.missLocked()逻辑大致为：
                        //   Map.misses++，如Map.misses大于一定值（Map.dirty长度），
                        //   则Map.dirty提升为Map.read，Map.dirty置nil，Map.misses置0
                        m.missLocked()
                }
                m.mu.Unlock()
        }
        if !ok {
                return nil, false
        }
        return e.load()
}
```

## write

```
func (m *Map) Store(key, value interface{}) {
    // 如果Map.read存在则尝试更新
    read, _ := m.read.Load().(readOnly)
    // entry.tryStore会在entry.p为expunged时返回false
    if e, ok := read.m[key]; ok && e.tryStore(&value) {
        return
    }

    // 此时`m.read`不存在或者已经被标记删除
    m.mu.Lock()
    read, _ = m.read.Load().(readOnly)
    if e, ok := read.m[key]; ok {
        // entry.unexpungeLocked()尝试将entry.p从expunged改为nil，并返回是否修改成功
        if e.unexpungeLocked() { 
            //修改成功，表明entry.p原为expunged，即Map.dirty中不存在key
            m.dirty[key] = e 
        }
        e.storeLocked(&value) //更新entry.p，指向value
    } else if e, ok := m.dirty[key]; ok { 
        // m.dirty存在这个键，更新
        e.storeLocked(&value)
    } else { //新键值
        if !read.amended { 
            m.dirtyLocked() //如果Map.dirty为nil，则Map.read中未删除的数据赋值给Map.dirty，且Map.read中为nil的entry置为expunged
            m.read.Store(readOnly{m: read.m, amended: true})
        }
        m.dirty[key] = newEntry(value) //将这个entry加入到m.dirty中
    }
    m.mu.Unlock()
}
```

## delete

```
func (m *Map) Delete(key interface{}) {
        read, _ := m.read.Load().(readOnly)
        e, ok := read.m[key]
        if !ok && read.amended {
                // Map.read中不存在且Map.dirty中存在多余的keys时
                // 从Map.dirty中删除key
                m.mu.Lock()
                read, _ = m.read.Load().(readOnly)
                e, ok = read.m[key]
                if !ok && read.amended {
                        delete(m.dirty, key)
                }
                m.mu.Unlock()
        }
        if ok {
                // entry.delete()会尝试将entry.p从指向数据改为nil
                e.delete()
        }
}
```

### range

```
func (m *Map) Range(f func(key, value interface{}) bool) {
        read, _ := m.read.Load().(readOnly)
        // 如果Map.dirty存在新keys，则提升Map.dirty
        if read.amended {
                m.mu.Lock()
                read, _ = m.read.Load().(readOnly)
                if read.amended {
                        read = readOnly{m: m.dirty}
                        m.read.Store(read)
                        m.dirty = nil
                        m.misses = 0
                }
                m.mu.Unlock()
        }

        for k, e := range read.m {
                v, ok := e.load()
                if !ok {
                        continue
                }
                if !f(k, v) {
                        break
                }
        }
}
```

参考[这里](https://studygolang.com/articles/10511)