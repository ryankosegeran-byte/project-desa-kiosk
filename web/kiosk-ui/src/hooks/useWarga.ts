import { useState, useCallback } from 'react';

export interface Warga {
  id: string;
  nik: string;
  rfid_uid?: string;
  nama: string;
  tempat_lahir?: string;
  tanggal_lahir?: string;
  jenis_kelamin?: string;
  alamat?: string;
  rt?: string;
  rw?: string;
  kelurahan?: string;
  kecamatan?: string;
  kabupaten?: string;
  provinsi?: string;
  agama?: string;
  status_kawin?: string;
  pekerjaan?: string;
  kewarganegaraan?: string;
  desa_id: string;
}

export function useWarga() {
  const [warga, setWarga] = useState<Warga | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  const apiBase = import.meta.env.DEV ? 'http://localhost:8080' : '';

  const clearWarga = useCallback(() => {
    setWarga(null);
    setError(null);
  }, []);

  const lookupByRFID = useCallback(async (uid: string): Promise<Warga | null> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/api/warga/rfid/${uid}`);
      if (!res.ok) {
        if (res.status === 404) {
          throw new Error("KTP Anda belum terdaftar di sistem desa. Silakan hubungi operator kantor desa.");
        }
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || "Gagal mengambil data warga");
      }
      const data: Warga = await res.json();
      setWarga(data);
      return data;
    } catch (err: any) {
      setError(err.message);
      setWarga(null);
      return null;
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  const lookupByNIK = useCallback(async (nik: string): Promise<Warga | null> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/api/warga/nik/${nik}`);
      if (!res.ok) {
        if (res.status === 404) {
          throw new Error("NIK tidak terdaftar. Silakan hubungi perangkat desa.");
        }
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || "Gagal mengambil data warga");
      }
      const data: Warga = await res.json();
      setWarga(data);
      return data;
    } catch (err: any) {
      setError(err.message);
      setWarga(null);
      return null;
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  const searchWarga = useCallback(async (query: string): Promise<Warga[]> => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${apiBase}/api/warga/search?q=${encodeURIComponent(query)}`);
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || "Gagal mencari warga");
      }
      const data: Warga[] = await res.json();
      return data;
    } catch (err: any) {
      setError(err.message);
      return [];
    } finally {
      setLoading(false);
    }
  }, [apiBase]);

  return {
    warga,
    loading,
    error,
    lookupByRFID,
    lookupByNIK,
    searchWarga,
    clearWarga,
    setWarga,
    setError
  };
}
