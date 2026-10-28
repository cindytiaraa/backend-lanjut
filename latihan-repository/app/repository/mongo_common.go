package repository

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/topology"
)

// Error dan helper yang dipakai bersama oleh seluruh repository MongoDB.
// ErrNotFound dan ErrDuplicate sudah didefinisikan di student_repository.go.
var ErrInvalidID = errors.New("format id tidak valid")

// IsUnavailable mengenali galat koneksi dan timeout MongoDB. Panggil hanya
// untuk error dari operasi MongoDB agar timeout database lain tidak tertukar.
func IsUnavailable(err error) bool {
	var selectionErr topology.ServerSelectionError
	return mongo.IsNetworkError(err) ||
		mongo.IsTimeout(err) ||
		errors.As(err, &selectionErr)
}

// now membulatkan waktu ke milidetik. Tipe date pada BSON hanya menyimpan
// sampai milidetik, sedangkan time.Time di Go menyimpan sampai nanodetik.
// Tanpa pembulatan, nilai yang dikirim ke client saat membuat dokumen akan
// berbeda dari nilai yang dibaca kembali dari database.
func now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

func parseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}
	return oid, nil
}
