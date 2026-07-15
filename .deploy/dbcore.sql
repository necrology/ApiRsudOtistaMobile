SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_KEY, ORDINAL_POSITION
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
AND TABLE_NAME IN (
'pasiens','registrasis','rawatinaps','folios','tagihans','pembayarans',
'hasillabs','rincian_hasillabs','order_lab','order_radiologi','radiologi_ekspertises','hasil_pemeriksaan',
'resep_note','resep_note_detail','penjualans','penjualandetails','masterobats',
'perawatan_icd10s','perawatan_icd9s','icd10s','icd9s',
'emr','emr_resume','emr_inap_pemeriksaans','emr_inap_medical_records','emr_konsuls','emr_farmasi',
'histori_kunjungan_irj','histori_rawatinap','histori_kunjungan_igd','histori_kunjungan_lab','histori_kunjungan_rad','histori_kunjungan_rm',
'dokumen_rekam_medis','resume_pasiens','data_seps','histori_seps','bpjs_rencana_kontrol',
'polis','pegawais','dokters','users'
)
ORDER BY FIELD(TABLE_NAME,
'pasiens','registrasis','rawatinaps','folios','tagihans','pembayarans',
'hasillabs','rincian_hasillabs','order_lab','order_radiologi','radiologi_ekspertises','hasil_pemeriksaan',
'resep_note','resep_note_detail','penjualans','penjualandetails','masterobats',
'perawatan_icd10s','perawatan_icd9s','icd10s','icd9s',
'emr','emr_resume','emr_inap_pemeriksaans','emr_inap_medical_records','emr_konsuls','emr_farmasi',
'histori_kunjungan_irj','histori_rawatinap','histori_kunjungan_igd','histori_kunjungan_lab','histori_kunjungan_rad','histori_kunjungan_rm',
'dokumen_rekam_medis','resume_pasiens','data_seps','histori_seps','bpjs_rencana_kontrol',
'polis','pegawais','dokters','users'), ORDINAL_POSITION;

SELECT TABLE_NAME, INDEX_NAME, GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX), NON_UNIQUE
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
AND TABLE_NAME IN ('pasiens','registrasis','rawatinaps','folios','hasillabs','rincian_hasillabs','order_lab','order_radiologi','radiologi_ekspertises','resep_note','resep_note_detail','penjualans','penjualandetails','perawatan_icd10s','perawatan_icd9s','emr','emr_resume')
GROUP BY TABLE_NAME, INDEX_NAME, NON_UNIQUE
ORDER BY TABLE_NAME, INDEX_NAME;