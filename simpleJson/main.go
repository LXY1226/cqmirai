package sJson

import "bytes"

func Unmarshal(b []byte) map[string][]byte {
	m := make(map[string][]byte)
	datas := bytes.Split(b[1:len(b)-1], []byte(","))
	for k := 0; k < len(datas); k++ {
		ke := datas[k][:bytes.IndexByte(datas[k], ':')]
		ke = bytes.TrimSpace(ke)
		key := string(ke[1 : len(ke)-1])
		if l := bytes.IndexByte(datas[k], '{'); l != -1 {
			if l := bytes.IndexByte(datas[k], '}'); l == -1 {
				j := k
				for ; j < len(datas); j++ {
					if l := bytes.IndexByte(datas[j], '}'); l != -1 {
						break
					}
				}

				m[key] = bytes.TrimSpace(append(datas[k][bytes.IndexByte(datas[k], ':')+1:], bytes.Join(datas[k+1:j+1], []byte(","))...))
				k = j
				continue
			}
		}
		if l := bytes.IndexByte(datas[k], '['); l != -1 {
			if l := bytes.IndexByte(datas[k], ']'); l == -1 {
				j := k
				for ; j < len(datas); j++ {
					if l := bytes.IndexByte(datas[j], ']'); l != -1 {
						break
					}
				}

				m[key] = bytes.TrimSpace(append(datas[k][bytes.IndexByte(datas[k], ':')+1:], bytes.Join(datas[k+1:j+1], []byte(","))...))
				k = j
				continue
			}
		}
		/*		if l := bytes.IndexByte(datas[k], '"'); l != -1 {
				if l := bytes.IndexByte(datas[k][l:], '"'); l == -1 {
					j := k
					for ; j < len(datas); j++ {
						if l := bytes.IndexByte(datas[j], ']'); l != -1 {
							break
						}
					}

					m[key] = bytes.TrimSpace(append(datas[k][bytes.IndexByte(datas[k], ':')+1:],bytes.Join(datas[k+1:j+1], []byte(","))...))
					k = j
					continue
				}
			}*/
		m[key] = datas[k]
		m[key] = bytes.TrimSpace(append(datas[k][bytes.IndexByte(datas[k], ':')+1:]))
	}
	return m
}
