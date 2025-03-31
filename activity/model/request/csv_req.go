package request

type CSVRecord struct {
	Data string `csv:"data"`
}

//
//// StringSlice 是一个自定义的字符串切片类型
//type StringSlice []string
//
//// BinaryMarshaler 将 StringSlice 编码为二进制数据
//func (ss StringSlice) BinaryMarshaler() ([][]byte, error) {
//	// 编码字符串切片的长度
//	var res [][]byte
//	// 遍历切片中的每个字符串，并编码它们的长度和内容
//	for _, s := range ss {
//		sByte := []byte(s)
//		res = append(res, sByte)
//	}
//
//	return res, nil
//}

//
//// UnmarshalBinary 从二进制数据中解码 StringSlice
//func (ss *StringSlice) UnmarshalBinary(data [][]byte) error {
//	// 重置当前切片
//	*ss = StringSlice{}
//
//	// 读取字符串切片的长度
//	reader := bytes.NewReader(data)
//	var length int32
//	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
//		if err == io.EOF {
//			return io.ErrUnexpectedEOF
//		}
//		return err
//	}
//
//	// 读取指定数量的字符串
//	for i := int32(0); i < length; i++ {
//		// 读取字符串的长度
//		var strLength int32
//		if err := binary.Read(reader, binary.LittleEndian, &strLength); err != nil {
//			return err
//		}
//
//		// 读取字符串的内容
//		strBytes := make([]byte, strLength)
//		if _, err := reader.Read(strBytes); err != nil {
//			return err
//		}
//
//		// 将字节切片转换为字符串并添加到切片中
//		*ss = append(*ss, string(strBytes))
//	}
//
//	return nil
//}
