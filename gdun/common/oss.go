package common

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"strings"
)

//----------------------------------------------------------------------------

type ObjectStorage struct {
	client *oss.Client
	bucket *oss.Bucket
}

func NewObjectStorage(endpoint string, bucketName string, keyID string, keySec string) (*ObjectStorage, error) {
	objs := new(ObjectStorage)

	var err error
	objs.client, err = oss.New(endpoint, keyID, keySec)
	if err != nil {
		return nil, err
	}

	objs.bucket, err = objs.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) UploadFile(key string, fileName string) error {
	return objs.bucket.PutObjectFromFile(key, fileName)
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) UploadString(key string, s string) error {
	return objs.bucket.PutObject(key, strings.NewReader(s))
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) UploadBuffer(key string, buf []byte) error {
	return objs.bucket.PutObject(key, bytes.NewReader(buf))
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) Exist(key string) (bool, error) {
	return objs.bucket.IsObjectExist(key)
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) Get(key string) ([]byte, error) {
	okay, err := objs.bucket.IsObjectExist(key)
	if err != nil {
		return nil, err
	}
	if !okay {
		return nil, ERR_NO_FILE
	}

	r, err := objs.bucket.GetObject(key)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return ioutil.ReadAll(r)
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) SaveToFile(key string, filename string) error {
	okay, err := objs.bucket.IsObjectExist(key)
	if err != nil {
		return err
	}
	if !okay {
		return ERR_NO_FILE
	}

	err = objs.bucket.GetObjectToFile(key, filename)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) Copy(from string, to string) error {
	_, err := objs.bucket.CopyObject(from, to)
	return err
}

//----------------------------------------------------------------------------

func (objs *ObjectStorage) Append(key string, buf []byte) error {
	// TODO:
	return nil
}

//----------------------------------------------------------------------------
