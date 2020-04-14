package bencoder
import(
	"bytes"
	"errors"
	"strconv"
	"unsafe"
)
func Decode(data []byte) (interface{}, error){
	return(&unmarshaler{
		data: data,
		length: len(data),
	}).unmarshal()
}
type unmarshaler struct{
	data []byte
	length int
	index int

}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (u *unmarshaler)unmarshal() (interface{}, error){
	switch u.data[u.index]{
	case 'i':
		u.index++
		index := bytes.IndexByte(u.data[u.index:], 'e')
		if index == -1{
			return nil, errors.New("Invalid integer field")
		}
		index += u.index
		integer, err := strconv.ParseInt(b2s(u.data[u.index:index]), 10, 64)
		if err != nil {
			return nil, err
		}
		u.index = index + 1
		return integer, nil
	case 'l':
		u.index++
		list := []interface{}{}
		for{
			if u.index == u.length{
				return nil, errors.New("Invalid list field")
			}
			if u.data[u.index] == 'e'{
				u.index++
				return list, nil
			}
			value, err := u.unmarshal()
			if err != nil{
				return nil, err
			}
			list = append(list, value)
		}
	case 'd':
		u.index++
		dictionary := map[string]interface{}{}
		for{
			if u.index == u.length{
				return nil, errors.New("Invalid dictionary field")
			}
			if u.data[u.index]  =='e' {
				u.index++
				return dictionary, nil
			}
			value, err := u.unmarshal()
			if err != nil{
				return nil, err
			}
			key, ok := value.([]byte)
			if !ok{
				return nil, errors.New("non-string dictionary key")
			}
			value, err = u.unmarshal()
			if err != nil {
				return nil, err
			}
			dictionary[b2s(key)] = value
		}
	default:
		index := bytes.IndexByte(u.data[u.index:], ':')
		if index == -1{
			return nil, errors.New("Invalid string field")
		}
		index += u.index
		stringLength, err := strconv.ParseInt(b2s(u.data[u.index:index]), 10, 64)
		if err != nil{
			return nil, err
		}
		index++
		endIndex := index + int(stringLength)
		if endIndex > u.length{
			return nil, errors.New("Invalid bencoded string")
		}
		value := u.data[index:endIndex]
		u.index = endIndex

		//fmt.Println(b2s(value))
		return value, nil
	}
}