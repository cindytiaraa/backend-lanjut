#!/usr/bin/env bash
# Bukti Tugas C.3: keluaran curl untuk setiap endpoint portofolio, termasuk
# kasus gagal, uji serentak batas 20 proyek, dan uji NoSQL injection.
#
# Prasyarat: aplikasi berjalan, PostgreSQL + MongoDB hidup, migration sudah
# dijalankan, dan ada satu student yang owner_id-nya adalah user yang login.
#
# Pakai:
#   BASE=http://localhost:3000/api/v1 USERNAME=mhs001 PASSWORD='Rahasia123!' \
#   STUDENT_ID=1 ALLOW_PORTFOLIO_MUTATION=1 ./scripts/uji_portofolio.sh | tee bukti_curl.txt
set -u

if [ "${ALLOW_PORTFOLIO_MUTATION:-}" != "1" ]; then
  echo "Skrip ini mengubah dan menghapus portofolio STUDENT_ID; gunakan data uji sekali pakai."
  echo "Set ALLOW_PORTFOLIO_MUTATION=1 untuk melanjutkan."
  exit 2
fi

BASE="${BASE:-http://localhost:3000/api/v1}"
USERNAME="${USERNAME:?isi USERNAME}"
PASSWORD="${PASSWORD:?isi PASSWORD}"
SID="${STUDENT_ID:?isi STUDENT_ID (id student milik user ini)}"
OTHER_SID="${OTHER_STUDENT_ID:-}"   # opsional: student milik orang lain, untuk uji 403

req() { # req <judul> <method> <path> [json]
  local title="$1" method="$2" path="$3" body="${4:-}"
  echo; echo "=== $title"
  echo "\$ $method $path ${body:+-d '$body'}"
  if [ -n "$body" ]; then
    curl -s -o /tmp/_out -w "HTTP %{http_code}\n" -X "$method" "$BASE$path" \
      -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$body"
  else
    curl -s -o /tmp/_out -w "HTTP %{http_code}\n" -X "$method" "$BASE$path" \
      -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json"
  fi
  cat /tmp/_out; echo
}

TOKEN=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" \
  | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')
[ -n "$TOKEN" ] || { echo "login gagal, periksa USERNAME/PASSWORD"; exit 1; }

echo "##### KASUS BERHASIL #####"
req "GET sebelum portofolio dibuat (404 PORTFOLIO_NOT_FOUND)" GET "/students/$SID/portfolio"
req "PUT membuat portofolio (201)" PUT "/students/$SID/portfolio" \
 '{"headline":"Backend Developer","bio":"Suka Go dan MongoDB","projects":[{"title":"API Kampus","technologies":["Go","PostgreSQL"],"url":"https://github.com/x/api-kampus","year":2025}],"certificates":[{"name":"Belajar Backend","issuer":"Dicoding","issued_at":"2025-03-01"}]}'
req "GET portofolio (200)" GET "/students/$SID/portfolio"
req "PUT kedua mengganti isi (200)" PUT "/students/$SID/portfolio" \
 '{"headline":"Backend Developer","bio":"Diganti","projects":[],"certificates":[]}'
req "POST proyek (201)" POST "/students/$SID/portfolio/projects" \
 '{"title":"Aplikasi Ulasan","technologies":["Go","MongoDB","Fiber"],"url":"https://github.com/x/ulasan","year":2026}'

PID=$(sed -n 's/.*"projects":\[{"id":"\([0-9a-f]*\)".*/\1/p' /tmp/_out)
req "DELETE proyek (204)" DELETE "/students/$SID/portfolio/projects/$PID"
req "Statistik teknologi (200)" GET "/portfolios/stats/technologies"

echo; echo "##### KASUS GAGAL #####"
req "Student tidak ada di PostgreSQL (404)" GET "/students/99999/portfolio"
req "Validasi gagal: headline kosong, tahun 1800, url javascript: (422)" PUT "/students/$SID/portfolio" \
 '{"headline":"","projects":[{"title":"X","year":1800,"url":"javascript:alert(1)"}]}'
req "Id proyek bukan ObjectID (400)" DELETE "/students/$SID/portfolio/projects/bukan-objectid"
req "Proyek tidak ada (404)" DELETE "/students/$SID/portfolio/projects/000000000000000000000000"
if [ -n "$OTHER_SID" ]; then
  req "Portofolio milik student lain (403)" PUT "/students/$OTHER_SID/portfolio" '{"headline":"Bajakan"}'
fi

echo; echo "##### NoSQL INJECTION (harus ditolak) #####"
req "Operator di dalam body (400)" PUT "/students/$SID/portfolio" '{"headline":{"$ne":""}}'
req "Operator di field proyek (400)" POST "/students/$SID/portfolio/projects" '{"title":{"$gt":""},"year":2025}'
req "Operator di path (400)" DELETE "/students/$SID/portfolio/projects/%7B%22%24ne%22%3Anull%7D"

echo; echo "##### BATAS 20 PROYEK, SERENTAK #####"
echo "Mengosongkan lalu mengisi 18 proyek..."
curl -s -o /dev/null -X PUT "$BASE/students/$SID/portfolio" -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"headline":"Uji batas","projects":[],"certificates":[]}'
for i in $(seq 1 18); do
  curl -s -o /dev/null -X POST "$BASE/students/$SID/portfolio/projects" \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d "{\"title\":\"Proyek $i\",\"technologies\":[\"go\"],\"year\":2025}"
done
echo "Mengirim 10 permintaan serentak (yang benar: 2 x 201, 8 x 409)..."
seq 1 10 | xargs -P 10 -I{} curl -s -o /dev/null -w "%{http_code}\n" -X POST \
  "$BASE/students/$SID/portfolio/projects" -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Serentak","technologies":["go"],"year":2026}' | sort | uniq -c
echo "Jumlah proyek akhir:"
curl -s "$BASE/students/$SID/portfolio" -H "Authorization: Bearer $TOKEN" \
  | grep -o '"title"' | wc -l
