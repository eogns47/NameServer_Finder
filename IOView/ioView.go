package ioview

import (
	"encoding/csv"
	"os"

	"github.com/pkg/errors"
)

func ReadCsv(target string) ([][]string, error) {
	// CSV 파일 열기
	file, err := os.Open("constants/" + target)
	if err != nil {
		return nil, errors.Wrap(err, "Open failed")
	}
	// 함수 종료시 파일 닫기
	defer file.Close()

	// CSV 리더 생성
	reader := csv.NewReader(file)
	_, err = reader.Read()

	// CSV 파일 읽기
	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "Read failed")
	}

	// 각 행의 데이터 출력
	return records, nil
}
