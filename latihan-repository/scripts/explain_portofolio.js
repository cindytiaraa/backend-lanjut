// Bukti Tugas C.3: index dipakai. Jalankan:
//   mongosh scripts/explain_portofolio.js
const d = db.getSiblingDB("praktikum_backend");
const sid = 1; // ganti dengan student_id yang sudah punya portofolio

print("== index pada student_portfolios ==");
printjson(d.student_portfolios.getIndexes().map(i => ({ name: i.name, key: i.key, unique: !!i.unique })));

print("\n== explain: GET /students/:id/portfolio ==");
const e = d.student_portfolios.find({ student_id: sid }).explain("executionStats");
const st = e.executionStats;
print("stage        : " + JSON.stringify(e.queryPlanner.winningPlan.inputStage ? e.queryPlanner.winningPlan.inputStage.stage : e.queryPlanner.winningPlan.stage));
print("index        : " + (e.queryPlanner.winningPlan.inputStage ? e.queryPlanner.winningPlan.inputStage.indexName : "-"));
print("diperiksa    : " + st.totalDocsExamined + " dokumen, kunci index: " + st.totalKeysExamined);
print("dikembalikan : " + st.nReturned);

print("\n== pembanding: tanpa index (hint $natural) ==");
const n = d.student_portfolios.find({ student_id: sid }).hint({ $natural: 1 }).explain("executionStats").executionStats;
print("diperiksa    : " + n.totalDocsExamined + " dokumen (COLLSCAN)");

print("\n== index unik menolak portofolio kedua ==");
try {
  d.student_portfolios.insertOne({ student_id: sid, headline: "ganda" });
} catch (err) { print(err.message); }
