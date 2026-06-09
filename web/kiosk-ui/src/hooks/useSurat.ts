import { useState, useCallback } from 'react';

export interface FieldDef {
  key: string;
  label: string;
  type: string; // text, textarea, number, date, select, radio, checkbox, repeater
  required: boolean;
  placeholder?: string;
  options?: string[];
  sub_fields?: FieldDef[];
}

export interface JenisSurat {
  id: string;
  kode: string;
  nama: string;
  deskripsi?: string;
  fields_schema: {
    fields: FieldDef[];
  };
  aktif: boolean;
  urutan: number;
}

export interface Surat {
  id: string;
  nomor_surat?: string;
  jenis_surat_id: string;
  jenis_surat_kode: string;
  jenis_surat_nama: string;
  warga_id?: string;
  nik_pemohon: string;
  nama_pemohon: string;
  data_surat: any;
  status: string; // DRAFT, PRINTED, SYNCED, FAILED
  pdf_path?: string;
  desa_id: string;
  created_at: string;
  printed_at?: string;
  synced: boolean;
}

export function useSurat() {
  const [jenisSuratList, setJenisSuratList] = useState<JenisSurat[]>([]);
  const [currentSurat, setCurrentSurat] = useState<Surat | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const apiBase = import.meta.env.DEV ? 'http://localhost:8080' : '';

  const fetchJenisSurat = useCallback(async (): Promise<JenisSurat[]> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/api/jenis-surat`);
      if (!res.ok) {
        throw new Error("Gagal memuat daftar jenis surat");
      }
      const data: JenisSurat[] = await res.json();
      setJenisSuratList(data);
      return data;
    } catch (err: any) {
      setError(err.message);
      return [];
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  const createSurat = useCallback(async (payload: {
    jenis_surat_id: string;
    warga_id?: string;
    nik_pemohon: string;
    nama_pemohon: string;
    data_surat: any;
  }): Promise<Surat | null> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/api/surat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || "Gagal membuat surat");
      }
      const data: Surat = await res.json();
      setCurrentSurat(data);
      return data;
    } catch (err: any) {
      setError(err.message);
      return null;
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  const printSurat = useCallback(async (id: string): Promise<Surat | null> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/api/surat/${id}/print`, {
        method: 'POST'
      });
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || "Gagal memproses cetak surat");
      }
      const data: Surat = await res.json();
      setCurrentSurat(data);
      return data;
    } catch (err: any) {
      setError(err.message);
      return null;
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  const fetchTemplateHTML = useCallback(async (jenisSuratID: string): Promise<string> => {
    try {
      const res = await fetch(`${apiBase}/api/template/${jenisSuratID}`);
      if (!res.ok) {
        throw new Error("Gagal mengambil template cetak");
      }
      const data = await res.json();
      return data.template_html || "";
    } catch (err) {
      console.error(err);
      return "";
    }
  }, [apiBase]);

  return {
    jenisSuratList,
    currentSurat,
    loading,
    error,
    fetchJenisSurat,
    createSurat,
    printSurat,
    fetchTemplateHTML,
    setCurrentSurat,
    setError
  };
}
