package ioview

import (
	"encoding/csv"
	"os"
)

func ReadCsv(target string) ([][]string, error){
	// CSV 파일 열기
	file, err := os.Open(target)
	if err != nil {
		return nil,err
	}
	defer file.Close()

	// CSV 리더 생성
	reader := csv.NewReader(file)
	_, err = reader.Read()

	// CSV 파일 읽기
	records, err := reader.ReadAll()
	if err != nil {
		return nil,err
	}

	// 각 행의 데이터 출력
	return records,nil
}

