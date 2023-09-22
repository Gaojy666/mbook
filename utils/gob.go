package utils

/*
* 系统gob封装
 */

import (
	"bytes"
	"encoding/gob"
)

//标准库gob是golang提供的“私有”的编解码方式，
//它的效率会比json，xml等更高，特别适合在Go语言程序间传递数据。

// 解码
func Decode(value string, r interface{}) error {
	//用于操作字节数据的缓冲区。它提供了一种方便的方式来读取、写入和操作字节数据，
	//同时还可以将缓冲区作为输入或输出进行传递。
	network := bytes.NewBuffer([]byte(value))
	// gob.NewDecoder创建一个解码器，将数据编码为字节流
	dec := gob.NewDecoder(network)
	//dec.Decode(r) 使用该解码器将字节流从 network 解码为变量 r 所指向的数据类型
	return dec.Decode(r)
}

// 编码
func Encode(value interface{}) (string, error) {
	//创建了一个 bytes.Buffer 对象 network
	//它一个实现了 io.Writer 和 io.Reader 接口的字节缓冲区。
	network := bytes.NewBuffer(nil)
	// gob.NewDecoder创建一个编码器，将数据编码为字节流
	enc := gob.NewEncoder(network)
	// 将 value 编码为字节流，并将字节流写入 network
	err := enc.Encode(value)
	if err != nil {
		return "", err
	}
	//它将字节缓冲区中的字节流转换为字符串形式
	return network.String(), nil
}
