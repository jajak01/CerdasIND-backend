import pandas as pd

# Define the columns for the enhanced template
# Column 0: Jenjang (SD/SMP/SMA)
# Column 1: Mapel (Matematika/IPA/dll)
# Column 2: Nama Bundle (Tryout 1/dll)
# Column 3: Waktu (Menit)
# Column 4: Tipe Soal (pilihan_ganda/isian_singkat)
# Column 5: Teks Soal
# Column 6: Kunci Jawaban
# Column 7: Pembahasan
# Column 8: Bobot Nilai
# Column 9-13: Pilihan A-E

data = [
    {
        "Jenjang": "SMA",
        "Mapel": "Biologi",
        "Nama Bundle": "Tryout Mitokondria",
        "Waktu Menit": 90,
        "Tipe Soal": "pilihan_ganda",
        "Teks Soal": "Pusat pernapasan sel adalah?",
        "Kunci Jawaban": "B",
        "Pembahasan": "Mitokondria berfungsi sebagai penghasil energi dan pusat pernapasan sel.",
        "Bobot Nilai": 10,
        "Pilihan A": "Ribosom",
        "Pilihan B": "Mitokondria",
        "Pilihan C": "Lisosom",
        "Pilihan D": "Badan Golgi",
        "Pilihan E": "Retikulum Endoplasma"
    },
    {
        "Jenjang": "SMA",
        "Mapel": "Biologi",
        "Nama Bundle": "Tryout Mitokondria",
        "Waktu Menit": 90,
        "Tipe Soal": "isian_singkat",
        "Teks Soal": "Cairan di dalam sel disebut?",
        "Kunci Jawaban": "Sitoplasma",
        "Pembahasan": "Sitoplasma adalah bagian sel yang terbungkus membran plasma.",
        "Bobot Nilai": 20,
        "Pilihan A": "",
        "Pilihan B": "",
        "Pilihan C": "",
        "Pilihan D": "",
        "Pilihan E": ""
    }
]

df = pd.DataFrame(data)

# Save to Excel
filename = "template_soal_lengkap.xlsx"
df.to_excel(filename, index=False, sheet_name="Sheet1")

print(f"Template Lengkap berhasil dibuat: {filename}")
