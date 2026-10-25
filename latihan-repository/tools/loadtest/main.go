// Alat uji beban Modul 9: membuktikan ada atau tidaknya race condition
// dengan mengirim puluhan request benar-benar serentak memakai goroutine.
//
// Folder tools/ bukan bagian aplikasi dan tidak ikut di-deploy.
//
// Cara pakai (server harus sedang berjalan di terminal lain):
//
//	go run ./tools/loadtest -mode crowd    # 60 mahasiswa berebut IF399 (kuota 5)
//	go run ./tools/loadtest -mode credits  # 10 mahasiswa x 12 kelas (batas 24 SKS)
//	go run ./tools/loadtest -mode swap     # 20 pasang pindah kelas IF101 <-> IF102
//	go run ./tools/loadtest -mode assign   # 30 penetapan serentak ke pembimbing berkuota 3 (D.3)
//	go run ./tools/loadtest -mode switch   # 20 pasang ganti pembimbing berlawanan arah (D.2)
//
// Untuk menghapus data uji sebelum berjalan, tambahkan -reset secara eksplisit.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-repository/app/model"
	"latihan-repository/config"
	"latihan-repository/database"
	"latihan-repository/helper"
)

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine dibuat, sehingga request benar-benar
// berangkat hampir bersamaan, bukan bergiliran sesuai urutan dibuatnya.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		// out bertipe chan<- (hanya kirim): goroutine ini hanya boleh
		// MENGIRIM hasil, tidak boleh membaca.
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	close(start) // buka gerbang
	wg.Wait()    // tunggu SEMUA goroutine selesai
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

type env struct {
	base   string
	pool   *pgxpool.Pool
	jwt    *helper.JWTManager
	client *http.Client
	userID map[string]int // username -> id
}

func main() {
	mode := flag.String("mode", "crowd", "crowd | credits | swap | assign | switch")
	base := flag.String("url", "http://localhost:3000", "alamat dasar server")
	resetData := flag.Bool("reset", false, "hapus dan reset data uji sebelum pengujian")
	noReset := flag.Bool("noreset", false, "alias kompatibilitas untuk tidak mereset data")
	flag.Parse()

	config.LoadEnv()
	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		fatal(err)
	}
	defer pool.Close()

	secret := config.GetEnv("JWT_SECRET", "")
	if secret == "" {
		fatal(fmt.Errorf("JWT_SECRET belum diisi di .env"))
	}
	// Alat ini mencetak token sendiri karena ia memegang JWT_SECRET.
	// Itulah sebabnya JWT_SECRET tidak boleh bocor: siapa pun yang
	// memegangnya dapat mencetak token atas nama siapa saja.
	issuer := config.GetEnv("JWT_ISSUER", "praktikum-backend")
	ttl := time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute

	e := &env{
		base: strings.TrimRight(*base, "/"),
		pool: pool,
		jwt:  helper.NewJWTManager(secret, issuer, ttl),
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        300,
				MaxIdleConnsPerHost: 300,
			},
		},
		userID: map[string]int{},
	}

	if err := e.loadStudents(ctx); err != nil {
		fatal(err)
	}
	if *resetData && !*noReset {
		if err := e.reset(ctx); err != nil {
			fatal(err)
		}
	}

	switch *mode {
	case "crowd":
		err = e.crowd(ctx)
	case "credits":
		err = e.credits(ctx)
	case "swap":
		err = e.swap(ctx)
	case "assign":
		err = e.assign(ctx)
	case "switch":
		err = e.switchSup(ctx)
	default:
		err = fmt.Errorf("mode tidak dikenal: %s", *mode)
	}
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "GAGAL:", err)
	os.Exit(1)
}

// ---------- persiapan ----------

func (e *env) loadStudents(ctx context.Context) error {
	rows, err := e.pool.Query(ctx, "SELECT id, username FROM users WHERE username LIKE 'mhs%'")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		e.userID[name] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(e.userID) < 60 {
		return fmt.Errorf("akun mhs001..mhs060 belum lengkap (ada %d): jalankan migrations/007_seed_loadtest.sql", len(e.userID))
	}
	return nil
}

// reset setara reset.sql, ditambah pemulihan kuota pembimbing.
func (e *env) reset(ctx context.Context) error {
	stmts := []string{
		`DELETE FROM enrollments WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%')`,
		`UPDATE courses c SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)`,
		`DELETE FROM supervisions WHERE student_id IN (SELECT id FROM students WHERE nim LIKE '2099%')`,
		`UPDATE supervisors SET quota = CASE WHEN name = 'Pembimbing Rebutan' THEN 3 ELSE 10 END`,
		`UPDATE supervisors s SET supervised_count = (SELECT COUNT(*) FROM supervisions v WHERE v.supervisor_id = s.id)`,
	}
	for _, q := range stmts {
		if _, err := e.pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("reset: %w", err)
		}
	}
	return nil
}

func (e *env) token(id int, username, role string) string {
	t, err := e.jwt.GenerateAccess(model.User{ID: id, Username: username, Role: role})
	if err != nil {
		fatal(err)
	}
	return t
}

func (e *env) studentToken(n int) string {
	name := fmt.Sprintf("mhs%03d", n)
	return e.token(e.userID[name], name, "user")
}

// adminToken memastikan akun admin uji ada, lalu mencetak tokennya.
func (e *env) adminToken(ctx context.Context) (string, int, error) {
	var id int
	err := e.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role)
		 VALUES ('loadadmin', 'loadadmin@student.unair.ac.id', '!', 'admin')
		 ON CONFLICT (LOWER(username)) DO UPDATE SET role = 'admin'
		 RETURNING id`).Scan(&id)
	if err != nil {
		return "", 0, err
	}
	return e.token(id, "loadadmin", "admin"), id, nil
}

// ---------- HTTP ----------

// call mengirim satu request dan mengembalikan label hasil, misalnya
// "201 OK" atau "409 COURSE_FULL".
func (e *env) call(method, path, token string, body any) string {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.base+path, rd)
	if err != nil {
		return "ERR " + err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return "ERR " + shorten(err.Error())
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return fmt.Sprintf("%d OK", resp.StatusCode)
	}
	var er struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(raw, &er)
	return fmt.Sprintf("%d %s", resp.StatusCode, er.Code)
}

func shorten(s string) string {
	if len(s) > 60 {
		return s[:60]
	}
	return s
}

// ---------- pelaporan ----------

func printTally(tally map[string]int, elapsed time.Duration) {
	fmt.Println("hasil request:")
	keys := make([]string, 0, len(tally))
	for k := range tally {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-28s %d\n", k, tally[k])
	}
	fmt.Printf("waktu total : %s\n", elapsed.Round(time.Millisecond))
}

func verifyTally(label string, actual, expected map[string]int) error {
	for result, count := range actual {
		if expected[result] != count {
			return fmt.Errorf("%s: hasil request tidak sesuai: aktual=%v, harapan=%v", label, actual, expected)
		}
	}
	for result, count := range expected {
		if actual[result] != count {
			return fmt.Errorf("%s: hasil request tidak sesuai: aktual=%v, harapan=%v", label, actual, expected)
		}
	}
	return nil
}

func verifyConsistent(label string, consistent bool) error {
	if !consistent {
		return fmt.Errorf("%s: invariant data tidak konsisten", label)
	}
	return nil
}

func verdict(ok bool) string {
	if ok {
		return "KONSISTEN"
	}
	return "TIDAK KONSISTEN"
}

func (e *env) deadlocks(ctx context.Context) (int64, error) {
	var n int64
	err := e.pool.QueryRow(ctx,
		"SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()").Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("membaca jumlah deadlock: %w", err)
	}
	return n, nil
}

// courseState mencetak dan memeriksa satu kelas: enrolled_count harus sama
// dengan jumlah baris enrollments DAN tidak melebihi kuota.
func (e *env) courseState(ctx context.Context, code string) (bool, error) {
	var quota, enrolled, rows int
	err := e.pool.QueryRow(ctx,
		`SELECT c.quota, c.enrolled_count, (SELECT COUNT(*) FROM enrollments WHERE course_id = c.id)
		   FROM courses c WHERE c.code = $1`, code).Scan(&quota, &enrolled, &rows)
	if err != nil {
		return false, err
	}
	ok := enrolled == rows && enrolled <= quota
	fmt.Printf("%s kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n",
		code, quota, enrolled, rows, verdict(ok))
	return ok, nil
}

func (e *env) courseID(ctx context.Context, code string) (int, error) {
	var id int
	err := e.pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = $1", code).Scan(&id)
	return id, err
}

// ---------- mode: crowd (C.1) ----------

func (e *env) crowd(ctx context.Context) error {
	id, err := e.courseID(ctx, "IF399")
	if err != nil {
		return err
	}
	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")

	tokens := make([]string, 60)
	for i := range tokens {
		tokens[i] = e.studentToken(i + 1)
	}

	begin := time.Now()
	tally := fire(60, func(i int) string {
		return e.call("POST", fmt.Sprintf("/api/v1/courses/%d/enroll", id), tokens[i], nil)
	})
	printTally(tally, time.Since(begin))
	if err := verifyTally("crowd", tally, map[string]int{"201 OK": 5, "409 COURSE_FULL": 55}); err != nil {
		return err
	}

	consistent, err := e.courseState(ctx, "IF399")
	if err != nil {
		return err
	}
	return verifyConsistent("crowd IF399", consistent)
}

// ---------- mode: credits (C.2) ----------

func (e *env) credits(ctx context.Context) error {
	rows, err := e.pool.Query(ctx,
		"SELECT id FROM courses WHERE code <> 'IF399' ORDER BY code")
	if err != nil {
		return err
	}
	var courseIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		courseIDs = append(courseIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	const students = 10
	per := len(courseIDs) // 12 kelas x 3 SKS = 36 SKS per mahasiswa bila lolos semua
	fmt.Printf("%d mahasiswa, masing-masing mendaftar ke %d kelas 3 SKS secara serentak (batas 24 SKS)...\n", students, per)

	tokens := make([]string, students)
	for i := range tokens {
		tokens[i] = e.studentToken(i + 1)
	}

	begin := time.Now()
	tally := fire(students*per, func(i int) string {
		s, c := i/per, i%per
		return e.call("POST", fmt.Sprintf("/api/v1/courses/%d/enroll", courseIDs[c]), tokens[s], nil)
	})
	printTally(tally, time.Since(begin))
	if err := verifyTally("credits", tally, map[string]int{"201 OK": 80, "409 CREDIT_LIMIT_EXCEEDED": 40}); err != nil {
		return err
	}

	all := true
	for i := 1; i <= students; i++ {
		var sum int
		if err := e.pool.QueryRow(ctx,
			`SELECT COALESCE(SUM(c.credits),0) FROM enrollments en JOIN courses c ON c.id = en.course_id
			  WHERE en.user_id = $1`, e.userID[fmt.Sprintf("mhs%03d", i)]).Scan(&sum); err != nil {
			return err
		}
		if sum > 24 {
			all = false
			fmt.Printf("mhs%03d memegang %d SKS (> 24)\n", i, sum)
		}
	}
	var drift int
	if err := e.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM courses c
		  WHERE enrolled_count <> (SELECT COUNT(*) FROM enrollments WHERE course_id = c.id)`).Scan(&drift); err != nil {
		return err
	}
	if drift > 0 {
		all = false
		fmt.Printf("%d kelas enrolled_count-nya tidak sama dengan jumlah baris\n", drift)
	}
	fmt.Printf("seluruh %d mahasiswa <= 24 SKS dan penghitung sesuai -> %s\n", students, verdict(all))
	return verifyConsistent("credits", all)
}

// ---------- mode: swap (C.3) ----------

func (e *env) swap(ctx context.Context) error {
	id101, err := e.courseID(ctx, "IF101")
	if err != nil {
		return err
	}
	id102, err := e.courseID(ctx, "IF102")
	if err != nil {
		return err
	}

	// Persiapan (tidak diukur): 20 mahasiswa di IF101, 20 di IF102.
	for i := 1; i <= 40; i++ {
		target := id101
		if i > 20 {
			target = id102
		}
		if r := e.call("POST", fmt.Sprintf("/api/v1/courses/%d/enroll", target), e.studentToken(i), nil); r != "201 OK" {
			return fmt.Errorf("persiapan mhs%03d gagal: %s", i, r)
		}
	}

	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")
	before, err := e.deadlocks(ctx)
	if err != nil {
		return err
	}

	begin := time.Now()
	tally := fire(40, func(i int) string {
		from, to := id101, id102
		if i >= 20 {
			from, to = id102, id101
		}
		return e.call("POST", "/api/v1/enrollments/swap", e.studentToken(i+1),
			model.SwapRequest{FromCourseID: from, ToCourseID: to})
	})
	elapsed := time.Since(begin)
	printTally(tally, elapsed)
	if err := verifyTally("swap", tally, map[string]int{"200 OK": 40}); err != nil {
		return err
	}

	time.Sleep(1500 * time.Millisecond) // beri waktu pg_stat_database menyegarkan diri
	deadlocksAfter, err := e.deadlocks(ctx)
	if err != nil {
		return err
	}
	deadlocks := deadlocksAfter - before
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", deadlocks)
	if deadlocks > 0 {
		return fmt.Errorf("swap: terdeteksi %d deadlock", deadlocks)
	}

	ok1, err := e.courseState(ctx, "IF101")
	if err != nil {
		return err
	}
	ok2, err := e.courseState(ctx, "IF102")
	if err != nil {
		return err
	}
	return verifyConsistent("swap", ok1 && ok2)
}

// ---------- persiapan pembimbing ----------

// seedSupervisionStudents memastikan 40 mahasiswa uji ada (nim 2099000001..40).
func (e *env) seedSupervisionStudents(ctx context.Context, ownerID int) ([]int, error) {
	ids := make([]int, 0, 40)
	for i := 1; i <= 40; i++ {
		var id int
		err := e.pool.QueryRow(ctx,
			`INSERT INTO students (nim, name, grade, is_active, owner_id)
			 VALUES ($1, $2, 80, TRUE, $3)
			 ON CONFLICT (LOWER(nim)) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			fmt.Sprintf("2099%06d", i), fmt.Sprintf("Mahasiswa Uji %02d", i), ownerID).Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (e *env) supervisorID(ctx context.Context, name string) (int, error) {
	var id int
	err := e.pool.QueryRow(ctx, "SELECT id FROM supervisors WHERE name = $1", name).Scan(&id)
	return id, err
}

func (e *env) supervisorState(ctx context.Context, name string) (bool, error) {
	var quota, count, rows int
	err := e.pool.QueryRow(ctx,
		`SELECT s.quota, s.supervised_count, (SELECT COUNT(*) FROM supervisions WHERE supervisor_id = s.id)
		   FROM supervisors s WHERE s.name = $1`, name).Scan(&quota, &count, &rows)
	if err != nil {
		return false, err
	}
	ok := count == rows && count <= quota
	fmt.Printf("%s kuota=%d supervised_count=%d baris supervisions=%d -> %s\n",
		name, quota, count, rows, verdict(ok))
	return ok, nil
}

// ---------- mode: assign (D.3) ----------

func (e *env) assign(ctx context.Context) error {
	admin, adminID, err := e.adminToken(ctx)
	if err != nil {
		return err
	}
	studentIDs, err := e.seedSupervisionStudents(ctx, adminID)
	if err != nil {
		return err
	}
	supID, err := e.supervisorID(ctx, "Pembimbing Rebutan")
	if err != nil {
		return err
	}

	fmt.Println("30 mahasiswa ditetapkan serentak ke pembimbing berkuota 3...")
	begin := time.Now()
	tally := fire(30, func(i int) string {
		return e.call("POST", fmt.Sprintf("/api/v1/supervisors/%d/assign", supID), admin,
			model.AssignRequest{StudentID: studentIDs[i]})
	})
	printTally(tally, time.Since(begin))
	if err := verifyTally("assign", tally, map[string]int{"201 OK": 3, "409 SUPERVISOR_FULL": 27}); err != nil {
		return err
	}

	consistent, err := e.supervisorState(ctx, "Pembimbing Rebutan")
	if err != nil {
		return err
	}
	return verifyConsistent("assign", consistent)
}

// ---------- mode: switch (D.2) ----------

func (e *env) switchSup(ctx context.Context) error {
	admin, adminID, err := e.adminToken(ctx)
	if err != nil {
		return err
	}
	studentIDs, err := e.seedSupervisionStudents(ctx, adminID)
	if err != nil {
		return err
	}
	alfa, err := e.supervisorID(ctx, "Pembimbing Alfa")
	if err != nil {
		return err
	}
	beta, err := e.supervisorID(ctx, "Pembimbing Beta")
	if err != nil {
		return err
	}
	// Beri ruang agar yang diuji murni urutan kunci, bukan kuota penuh.
	if _, err := e.pool.Exec(ctx, "UPDATE supervisors SET quota = 100 WHERE id IN ($1, $2)", alfa, beta); err != nil {
		return err
	}

	// Persiapan: 20 mahasiswa pertama di Alfa, 20 berikutnya di Beta.
	for i := 0; i < 40; i++ {
		target := alfa
		if i >= 20 {
			target = beta
		}
		if r := e.call("POST", fmt.Sprintf("/api/v1/supervisors/%d/assign", target), admin,
			model.AssignRequest{StudentID: studentIDs[i]}); r != "201 OK" {
			return fmt.Errorf("persiapan mahasiswa %d gagal: %s", i+1, r)
		}
	}

	fmt.Println("20 pasang mahasiswa ganti pembimbing berlawanan arah secara serentak...")
	before, err := e.deadlocks(ctx)
	if err != nil {
		return err
	}
	begin := time.Now()
	tally := fire(40, func(i int) string {
		to := beta
		if i >= 20 {
			to = alfa
		}
		return e.call("POST", "/api/v1/supervisions/switch", admin,
			model.SwitchSupervisionRequest{StudentID: studentIDs[i], ToSupervisorID: to})
	})
	printTally(tally, time.Since(begin))
	if err := verifyTally("switch", tally, map[string]int{"200 OK": 40}); err != nil {
		return err
	}

	time.Sleep(1500 * time.Millisecond)
	deadlocksAfter, err := e.deadlocks(ctx)
	if err != nil {
		return err
	}
	deadlocks := deadlocksAfter - before
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", deadlocks)
	if deadlocks > 0 {
		return fmt.Errorf("switch: terdeteksi %d deadlock", deadlocks)
	}

	okAlfa, err := e.supervisorState(ctx, "Pembimbing Alfa")
	if err != nil {
		return err
	}
	okBeta, err := e.supervisorState(ctx, "Pembimbing Beta")
	if err != nil {
		return err
	}
	return verifyConsistent("switch", okAlfa && okBeta)
}
