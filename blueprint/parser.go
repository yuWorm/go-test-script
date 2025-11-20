package blueprint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Parser 负责解析蓝图 JSON
type Parser struct {
	// 可以添加配置选项
	strictMode bool // 严格模式：更严格的验证
}

// NewParser 创建一个新的解析器
func NewParser() *Parser {
	return &Parser{
		strictMode: true,
	}
}

// ParseFromBytes 从字节数组解析蓝图
func (p *Parser) ParseFromBytes(data []byte) (*Blueprint, error) {
	var bp Blueprint
	if err := json.Unmarshal(data, &bp); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint JSON: %w", err)
	}

	// 初始化内部字段
	bp.nodeMap = make(map[string]*Node)
	for _, node := range bp.Nodes {
		bp.nodeMap[node.ID] = node
	}

	// 验证蓝图
	if err := bp.Validate(); err != nil {
		return nil, fmt.Errorf("blueprint validation failed: %w", err)
	}

	return &bp, nil
}

// ParseFromReader 从 io.Reader 解析蓝图
func (p *Parser) ParseFromReader(reader io.Reader) (*Blueprint, error) {
	var bp Blueprint
	decoder := json.NewDecoder(reader)

	// 严格模式下不允许未知字段
	if p.strictMode {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(&bp); err != nil {
		return nil, fmt.Errorf("failed to decode blueprint JSON: %w", err)
	}

	// 初始化内部字段
	bp.nodeMap = make(map[string]*Node)
	for _, node := range bp.Nodes {
		bp.nodeMap[node.ID] = node
	}

	// 验证蓝图
	if err := bp.Validate(); err != nil {
		return nil, fmt.Errorf("blueprint validation failed: %w", err)
	}

	return &bp, nil
}

// ParseFromFile 从文件解析蓝图
func (p *Parser) ParseFromFile(filename string) (*Blueprint, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	return p.ParseFromReader(file)
}

// ParseFromString 从字符串解析蓝图
func (p *Parser) ParseFromString(jsonStr string) (*Blueprint, error) {
	return p.ParseFromBytes([]byte(jsonStr))
}

// Serializer 负责序列化蓝图为 JSON
type Serializer struct {
	indent bool // 是否格式化输出
}

// NewSerializer 创建一个新的序列化器
func NewSerializer(indent bool) *Serializer {
	return &Serializer{
		indent: indent,
	}
}

// ToBytes 序列化蓝图为字节数组
func (s *Serializer) ToBytes(bp *Blueprint) ([]byte, error) {
	if s.indent {
		return json.MarshalIndent(bp, "", "  ")
	}
	return json.Marshal(bp)
}

// ToString 序列化蓝图为字符串
func (s *Serializer) ToString(bp *Blueprint) (string, error) {
	bytes, err := s.ToBytes(bp)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ToFile 序列化蓝图到文件
func (s *Serializer) ToFile(bp *Blueprint, filename string) error {
	data, err := s.ToBytes(bp)
	if err != nil {
		return fmt.Errorf("failed to serialize blueprint: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// ToWriter 序列化蓝图到 io.Writer
func (s *Serializer) ToWriter(bp *Blueprint, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	if s.indent {
		encoder.SetIndent("", "  ")
	}

	if err := encoder.Encode(bp); err != nil {
		return fmt.Errorf("failed to encode blueprint: %w", err)
	}

	return nil
}
