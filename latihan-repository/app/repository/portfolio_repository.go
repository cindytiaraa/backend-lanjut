package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"latihan-repository/app/model"
)

var (
	// ErrPortfolioNotFound: mahasiswanya ada, tetapi belum punya portofolio.
	ErrPortfolioNotFound = errors.New("portofolio belum dibuat")
	ErrProjectNotFound   = errors.New("proyek tidak ditemukan")
	ErrProjectLimit      = errors.New("jumlah proyek sudah mencapai batas")
)

// MaxProjectsPerPortfolio membatasi pertumbuhan array projects (batas dokumen
// MongoDB 16 MB). Dijaga secara atomik di dalam filter, bukan di kode Go.
const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error)
	// Replace membuat atau mengganti portofolio. Nilai bool true berarti
	// dokumen baru dibuat (bukan diganti).
	Replace(ctx context.Context, p model.Portfolio) (model.Portfolio, bool, error)
	AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error)
	RemoveProject(ctx context.Context, studentID int, projectID string) error
	TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyStat, error)
}

// Dokumen sebagaimana tersimpan di MongoDB (lihat catatan pada reviewDocument).
type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	URL          string        `bson:"url"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	ID       bson.ObjectID `bson:"_id"`
	Name     string        `bson:"name"`
	Issuer   string        `bson:"issuer"`
	IssuedAt time.Time     `bson:"issued_at"`
}

func newProjectDocument(p model.Project) projectDocument {
	techs := p.Technologies
	if techs == nil {
		techs = []string{}
	}
	return projectDocument{
		ID: bson.NewObjectID(), Title: p.Title, Technologies: techs,
		URL: p.URL, Year: p.Year,
	}
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, model.Project{
			ID: p.ID.Hex(), Title: p.Title, Technologies: techs, URL: p.URL, Year: p.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID: c.ID.Hex(), Name: c.Name, Issuer: c.Issuer,
			IssuedAt: c.IssuedAt.UTC().Format("2006-01-02"),
		})
	}

	return model.Portfolio{
		ID: d.ID.Hex(), StudentID: d.StudentID, Headline: d.Headline, Bio: d.Bio,
		Projects: projects, Certificates: certs,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type portfolioMongoRepository struct{ coll *mongo.Collection }

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// EnsureIndexes membuat index unik student_id: "satu mahasiswa satu
// portofolio" dijaga DATABASE, sehingga dua PUT pertama yang datang serentak
// tidak dapat menghasilkan dua dokumen.
//
// Index inilah satu-satunya yang dibutuhkan: seluruh endpoint mencari dokumen
// lewat student_id. Pipeline statistik teknologi memang membaca seluruh
// collection, dan index tidak membantunya, jadi tidak ada index tambahan
// (setiap index memperlambat penulisan).
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("student_unique"),
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrPortfolioNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// Replace melakukan upsert: membuat dokumen bila belum ada, mengganti isinya
// bila sudah ada. created_at hanya diisi saat dokumen dibuat ($setOnInsert).
//
// Dua PUT pertama yang serentak dapat sama-sama mencoba menyisipkan; yang
// kalah ditolak index unik dengan duplicate key. Itu bukan kegagalan sungguhan,
// jadi operasi diulang sekali, dan kali ini dokumennya sudah ada sehingga
// hasilnya berupa penggantian.
func (r *portfolioMongoRepository) Replace(ctx context.Context, p model.Portfolio) (model.Portfolio, bool, error) {
	projects := make([]projectDocument, 0, len(p.Projects))
	for _, pr := range p.Projects {
		projects = append(projects, newProjectDocument(pr))
	}

	certs := make([]certificateDocument, 0, len(p.Certificates))
	for _, c := range p.Certificates {
		issued, err := time.Parse("2006-01-02", c.IssuedAt)
		if err != nil {
			return model.Portfolio{}, false, fmt.Errorf("tanggal sertifikat %q: %w", c.IssuedAt, err)
		}
		certs = append(certs, certificateDocument{
			ID: bson.NewObjectID(), Name: c.Name, Issuer: c.Issuer, IssuedAt: issued.UTC(),
		})
	}

	t := now()
	update := bson.M{
		"$set": bson.M{
			"headline":     p.Headline,
			"bio":          p.Bio,
			"projects":     projects,
			"certificates": certs,
			"updated_at":   t,
		},
		"$setOnInsert": bson.M{"student_id": p.StudentID, "created_at": t},
	}

	var res *mongo.UpdateResult
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		res, err = r.coll.UpdateOne(ctx,
			bson.M{"student_id": p.StudentID}, update,
			options.UpdateOne().SetUpsert(true))
		if err == nil || !mongo.IsDuplicateKeyError(err) {
			break
		}
	}
	if err != nil {
		return model.Portfolio{}, false, fmt.Errorf("menyimpan portofolio: %w", err)
	}

	saved, err := r.FindByStudent(ctx, p.StudentID)
	if err != nil {
		return model.Portfolio{}, false, err
	}
	return saved, res.UpsertedCount > 0, nil
}

// AddProject menambah satu proyek dengan $push, dibatasi MaxProjectsPerPortfolio.
//
// Syarat "projects.19 tidak ada" di dalam filter berarti array belum berisi 20
// elemen. Pemeriksaan batas dan penambahan terjadi dalam SATU operasi atomik,
// sama seperti pola "replies.49" pada ulasan.
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error) {
	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)

	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
		bson.M{
			"$push": bson.M{"projects": newProjectDocument(p)},
			"$set":  bson.M{"updated_at": now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Tidak ada yang cocok: portofolionya belum ada, atau proyek sudah penuh.
	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

// RemoveProject menghapus satu proyek dengan $pull.
//
// Syarat "projects._id = projectID" ada di dalam filter, jadi filter hanya
// cocok bila proyeknya memang ada. Bila tidak ada yang cocok, dibedakan antara
// "portofolionya tidak ada" dan "proyeknya tidak ada".
func (r *portfolioMongoRepository) RemoveProject(ctx context.Context, studentID int, projectID string) error {
	pid, err := parseID(projectID)
	if err != nil {
		return err
	}

	res, err := r.coll.UpdateOne(ctx,
		bson.M{"student_id": studentID, "projects._id": pid},
		bson.M{
			"$pull": bson.M{"projects": bson.M{"_id": pid}},
			"$set":  bson.M{"updated_at": now()},
		},
	)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount > 0 {
		return nil
	}

	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return ErrPortfolioNotFound
	}
	return ErrProjectNotFound
}

// TopTechnologies: aggregation pipeline dengan DUA kali $unwind.
//
//	{projects:[{technologies:[go,sql]}, {technologies:[go]}]}
//	$unwind "$projects"              -> 2 dokumen (satu per proyek)
//	$unwind "$projects.technologies" -> 3 dokumen (satu per pasangan proyek-teknologi)
//	$group per teknologi             -> go: 2 proyek, sql: 1 proyek
//
// "students" menghitung mahasiswa BERBEDA lewat $addToSet, sehingga satu
// mahasiswa dengan lima proyek Go tetap dihitung satu mahasiswa.
func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyStat, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":      "$projects.technologies",
			"projects": bson.M{"$sum": 1},
			"students": bson.M{"$addToSet": "$student_id"},
		}}},
		{{Key: "$project", Value: bson.M{
			"projects": 1,
			"students": bson.M{"$size": "$students"},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "projects", Value: -1},
			{Key: "students", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung statistik teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Technology string `bson:"_id"`
		Projects   int    `bson:"projects"`
		Students   int    `bson:"students"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("membaca statistik teknologi: %w", err)
	}

	result := make([]model.TechnologyStat, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.TechnologyStat{
			Technology: row.Technology, Projects: row.Projects, Students: row.Students,
		})
	}
	return result, nil
}
