package storagex

import "cloud.google.com/go/storage"

func SetItNext(b Bucket, f func(it *storage.ObjectIterator) (*storage.ObjectAttrs, error)) {
	b.itNext = f
}

func NewBucketWithIt(bucket *storage.BucketHandle, it f func(it *storage.ObjectIterator) (*storage.ObjectAttrs, error)) {

}
