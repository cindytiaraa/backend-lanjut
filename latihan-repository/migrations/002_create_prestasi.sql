CREATE TABLE IF NOT EXISTS prestasi (
    id_prestasi SERIAL PRIMARY KEY,
    id_student INTEGER NOT NULL,
    nama_prestasi VARCHAR(255) NOT NULL,
    juara VARCHAR(100) NOT NULL,

    CONSTRAINT fk_prestasi_student
        FOREIGN KEY (id_student)
        REFERENCES students(id)
        ON DELETE CASCADE
);
