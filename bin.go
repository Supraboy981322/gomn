package gomn

import (
	"os"
	"encoding/gob"
)

func WrBin(gomn Map, file string) error {
	var fi *os.File
	var err error

	fi, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { return err }
	defer fi.Close()

	enc := gob.NewEncoder(fi)

	if err = enc.Encode(gomn); err != nil {
		return err
	}

	return nil
}

func ReadBin(file string) (Map, error) {
	fi, err := os.Open(file)
	if err != nil { return make(Map), err }
	defer fi.Close()

	var gomn Map
	dec := gob.NewDecoder(fi)
	if err = dec.Decode(&gomn); err != nil {
		return make(Map), err
	}

	return gomn, nil
}
