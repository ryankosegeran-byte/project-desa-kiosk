import React, { useEffect, useMemo, useState } from "react";
import { request, getUser } from "../lib/api";
import LinkRFIDModal from "./LinkRFIDModal";
import {
  useReactTable,
  getCoreRowModel,
  getPaginationRowModel,
  getFilteredRowModel,
  flexRender,
  createColumnHelper,
  type ColumnDef,
} from "@tanstack/react-table";

interface Warga {
  id: string;
  nik: string;
  nama: string;
  tempat_lahir?: string;
  tanggal_lahir?: string;
  jenis_kelamin?: string;
  alamat?: string;
  rt?: string;
  rw?: string;
  rfid_uid?: string;
  status?: string;
  draft_token?: string;
  desa_id?: string;
  deleted_at?: string;
}

interface Desa {
  id: string;
  nama: string;
}

type Tab = "aktif" | "hantu";

export default function WargaList() {
  const user = getUser();
  const isSuperadmin = user?.role === "superadmin";

  const [tab, setTab] = useState<Tab>("aktif");
  const [warga, setWarga] = useState<Warga[]>([]);
  const [desaList, setDesaList] = useState<Desa[]>([]);
  const [desaFilter, setDesaFilter] = useState("");
  const [globalFilter, setGlobalFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // RFID linking modal state
  const [linkingWarga, setLinkingWarga] = useState<Warga | null>(null);

  // Delete state
  const [deleteConfirm, setDeleteConfirm] = useState<Warga | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [hardDeleteConfirm, setHardDeleteConfirm] = useState<Warga | null>(null);
  const [hardDeleteLoading, setHardDeleteLoading] = useState(false);

  async function loadWarga() {
    setLoading(true);
    setError("");
    try {
      const params = new URLSearchParams();
      if (tab === "hantu") params.set("deleted", "true");
      if (isSuperadmin && desaFilter) params.set("desa_id", desaFilter);
      const qs = params.toString();
      const data = await request(`/api/warga${qs ? `?${qs}` : ""}`);
      setWarga(Array.isArray(data) ? data : []);
    } catch (err: any) {
      setError(err.message || "Gagal mengambil data warga.");
    } finally {
      setLoading(false);
    }
  }

  // Load warga whenever tab or desa filter changes
  useEffect(() => {
    loadWarga();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, desaFilter]);

  // Load desa list once (superadmin filter dropdown)
  useEffect(() => {
    if (!isSuperadmin) return;
    request("/api/desa")
      .then((data: any) => setDesaList(Array.isArray(data) ? data : []))
      .catch(() => setDesaList([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDeleteWarga = async () => {
    if (!deleteConfirm) return;
    setDeleteLoading(true);
    try {
      await request(`/api/warga/${deleteConfirm.id}`, { method: "DELETE" });
      await loadWarga();
      setDeleteConfirm(null);
    } catch (err: any) {
      alert(err.message || "Gagal menghapus warga.");
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleHardDeleteWarga = async () => {
    if (!hardDeleteConfirm) return;
    setHardDeleteLoading(true);
    try {
      await request(`/api/warga/${hardDeleteConfirm.id}/permanent`, { method: "DELETE" });
      await loadWarga();
      setHardDeleteConfirm(null);
    } catch (err: any) {
      alert(err.message || "Gagal menghapus permanen.");
    } finally {
      setHardDeleteLoading(false);
    }
  };

  const desaName = (id?: string) => desaList.find((d) => d.id === id)?.nama || "-";

  // ----- Table columns -----
  const columnHelper = createColumnHelper<Warga>();
  const columns = useMemo<ColumnDef<Warga, any>[]>(() => {
    const base: ColumnDef<Warga, any>[] = [
      columnHelper.accessor("nik", {
        header: "NIK",
        cell: (info) => info.getValue() || <em style={{ color: "var(--text-muted)" }}>belum diisi</em>,
      }),
      columnHelper.accessor("nama", {
        header: "Nama Lengkap",
        cell: (info) => info.getValue() || <em style={{ color: "var(--text-muted)" }}>belum diisi</em>,
      }),
      columnHelper.display({
        id: "ttl",
        header: "Tempat / Tgl Lahir",
        cell: ({ row }) => `${row.original.tempat_lahir || "-"}, ${row.original.tanggal_lahir || "-"}`,
      }),
      columnHelper.accessor("jenis_kelamin", {
        header: "Jenis Kelamin",
        cell: (info) => (info.getValue() === "L" ? "Laki-laki" : info.getValue() === "P" ? "Perempuan" : "-"),
      }),
      columnHelper.display({
        id: "alamat",
        header: "Alamat",
        cell: ({ row }) =>
          `${row.original.alamat || ""} ${row.original.rt ? `RT ${row.original.rt}` : ""} ${row.original.rw ? `/ RW ${row.original.rw}` : ""}`.trim() || "-",
      }),
      columnHelper.accessor("rfid_uid", {
        header: "Kartu RFID",
        cell: (info) =>
          info.getValue() ? (
            <span className="badge badge-success">{info.getValue()}</span>
          ) : (
            <span className="badge badge-warning">Belum Ada</span>
          ),
      }),
    ];

    if (isSuperadmin) {
      base.push(
        columnHelper.display({
          id: "desa",
          header: "Desa",
          cell: ({ row }) => desaName(row.original.desa_id),
        }),
      );
    }

    if (tab === "aktif") {
      base.push(
        columnHelper.display({
          id: "aksi",
          header: "Aksi",
          cell: ({ row }) => {
            const w = row.original;
            return (
              <div style={{ display: "flex", gap: "6px", flexWrap: "wrap" }}>
                {w.status === "draft" && w.draft_token ? (
                  <a href={`/warga/draft?token=${w.draft_token}`} className="btn btn-primary" style={{ padding: "6px 12px", fontSize: "13px", textDecoration: "none" }}>
                    Lanjutkan
                  </a>
                ) : (
                  <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => setLinkingWarga(w)}>
                    Tautkan Kartu
                  </button>
                )}
                <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px", color: "var(--danger)", borderColor: "var(--danger)" }} onClick={() => setDeleteConfirm(w)}>
                  Hapus
                </button>
              </div>
            );
          },
        }),
      );
    } else {
      base.push(
        columnHelper.accessor("deleted_at", {
          header: "Dihapus Pada",
          cell: (info) => (info.getValue() ? new Date(info.getValue() as string).toLocaleString("id-ID") : "-"),
        }),
        columnHelper.display({
          id: "aksi-hantu",
          header: "Aksi",
          cell: ({ row }) => (
            <button
              className="btn btn-secondary"
              style={{ padding: "6px 12px", fontSize: "13px", color: "var(--danger)", borderColor: "var(--danger)" }}
              onClick={() => setHardDeleteConfirm(row.original)}
            >
              Hapus Permanen
            </button>
          ),
        }),
      );
    }

    return base;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isSuperadmin, tab, desaList]);

  const table = useReactTable({
    data: warga,
    columns,
    state: { globalFilter },
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const q = String(filterValue).toLowerCase();
      const w = row.original;
      return (w.nama || "").toLowerCase().includes(q) || (w.nik || "").includes(q);
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: { pagination: { pageSize: 10 } },
  });

  return (
    <div>
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px", flexWrap: "wrap", gap: "12px" }}>
        <div>
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Data Warga Desa</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Kelola profil kependudukan warga desa yang sinkron dengan kiosk pelayanan mandiri.
          </p>
        </div>
        <a href="/warga/register" className="btn btn-primary">+ Registrasi Warga Baru</a>
      </div>

      {/* Tabs */}
      <div style={{ display: "flex", gap: "8px", marginBottom: "16px", borderBottom: "1px solid var(--border-color)" }}>
        {([
          { key: "aktif", label: "Data Aktif" },
          { key: "hantu", label: "Data Terhapus" },
        ] as { key: Tab; label: string }[]).map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            style={{
              padding: "10px 18px",
              background: "transparent",
              border: "none",
              borderBottom: tab === t.key ? "2px solid var(--primary)" : "2px solid transparent",
              color: tab === t.key ? "var(--primary)" : "var(--text-muted)",
              fontWeight: 600,
              cursor: "pointer",
              fontSize: "14px",
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          {error}
        </div>
      )}

      {/* Filters */}
      <div className="glass-card" style={{ padding: "16px", marginBottom: "24px", display: "flex", gap: "12px", flexWrap: "wrap", alignItems: "center" }}>
        <input
          type="text"
          className="form-control"
          placeholder="Cari nama atau NIK warga..."
          value={globalFilter}
          onChange={(e) => setGlobalFilter(e.target.value)}
          style={{ flex: "1 1 280px" }}
        />
        {isSuperadmin && (
          <select className="form-control" value={desaFilter} onChange={(e) => setDesaFilter(e.target.value)} style={{ flex: "0 1 240px" }}>
            <option value="">Semua Desa</option>
            {desaList.map((d) => (
              <option key={d.id} value={d.id}>{d.nama}</option>
            ))}
          </select>
        )}
        <select
          className="form-control"
          value={table.getState().pagination.pageSize}
          onChange={(e) => table.setPageSize(Number(e.target.value))}
          style={{ flex: "0 0 130px" }}
        >
          {[10, 25, 50, 100].map((n) => (
            <option key={n} value={n}>{n} / halaman</option>
          ))}
        </select>
      </div>

      {/* Table */}
      {loading ? (
        <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "300px" }}>
          <div className="spinner"></div>
        </div>
      ) : (
        <>
          <div className="table-container">
            <table className="premium-table">
              <thead>
                {table.getHeaderGroups().map((hg) => (
                  <tr key={hg.id}>
                    {hg.headers.map((h) => (
                      <th key={h.id}>{flexRender(h.column.columnDef.header, h.getContext())}</th>
                    ))}
                  </tr>
                ))}
              </thead>
              <tbody>
                {table.getRowModel().rows.map((row) => (
                  <tr key={row.id} style={row.original.status === "draft" ? { opacity: 0.75 } : {}}>
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</td>
                    ))}
                  </tr>
                ))}
                {table.getRowModel().rows.length === 0 && (
                  <tr>
                    <td colSpan={columns.length} style={{ textAlign: "center", color: "var(--text-muted)", padding: "24px" }}>
                      {tab === "hantu" ? "Tidak ada data terhapus" : "Warga tidak ditemukan"}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination controls */}
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: "16px", flexWrap: "wrap", gap: "12px" }}>
            <span style={{ color: "var(--text-muted)", fontSize: "13px" }}>
              Menampilkan {table.getRowModel().rows.length} dari {table.getFilteredRowModel().rows.length} data
            </span>
            <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
              <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => table.setPageIndex(0)} disabled={!table.getCanPreviousPage()}>{"<<"}</button>
              <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => table.previousPage()} disabled={!table.getCanPreviousPage()}>{"<"}</button>
              <span style={{ fontSize: "13px", fontWeight: 600 }}>
                Hal {table.getState().pagination.pageIndex + 1} / {table.getPageCount() || 1}
              </span>
              <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => table.nextPage()} disabled={!table.getCanNextPage()}>{">"}</button>
              <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => table.setPageIndex(table.getPageCount() - 1)} disabled={!table.getCanNextPage()}>{">>"}</button>
            </div>
          </div>
        </>
      )}

      {/* RFID Link Modal */}
      {linkingWarga && (
        <LinkRFIDModal
          warga={linkingWarga}
          desaIdHint={isSuperadmin ? desaFilter : ""}
          onClose={() => setLinkingWarga(null)}
          onLinked={() => { setLinkingWarga(null); loadWarga(); }}
        />
      )}

      {/* Delete Confirmation Modal */}
      {deleteConfirm && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "var(--overlay)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "420px", width: "100%", padding: "30px", textAlign: "center" }}>
            <div style={{ width: "56px", height: "56px", borderRadius: "50%", background: "rgba(239,68,68,0.12)", display: "flex", alignItems: "center", justifyContent: "center", margin: "0 auto 16px", color: "var(--danger)", fontSize: "28px" }}>!</div>
            <h3 style={{ fontSize: "18px", fontWeight: "700", marginBottom: "8px" }}>Hapus Data Warga?</h3>
            <p style={{ color: "var(--text-muted)", fontSize: "14px", marginBottom: "20px" }}>
              Data <strong>{deleteConfirm.nama || "tanpa nama"}</strong> ({deleteConfirm.nik || "tanpa NIK"}) akan dihapus. NIK & kartu akan dibebaskan agar bisa didaftarkan ulang.
            </p>
            <div style={{ display: "flex", justifyContent: "center", gap: "12px" }}>
              <button className="btn btn-secondary" onClick={() => setDeleteConfirm(null)} disabled={deleteLoading}>Batal</button>
              <button className="btn btn-primary" style={{ background: "var(--danger)", borderColor: "var(--danger)" }} onClick={handleDeleteWarga} disabled={deleteLoading}>
                {deleteLoading ? "Menghapus..." : "Ya, Hapus"}
              </button>
            </div>
          </div>
        </div>
      )}
      {/* Hard Delete Confirmation Modal */}
      {hardDeleteConfirm && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "var(--overlay)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "440px", width: "100%", padding: "30px", textAlign: "center" }}>
            <div style={{ width: "56px", height: "56px", borderRadius: "50%", background: "rgba(239,68,68,0.15)", display: "flex", alignItems: "center", justifyContent: "center", margin: "0 auto 16px", color: "var(--danger)", fontSize: "28px" }}>!</div>
            <h3 style={{ fontSize: "18px", fontWeight: "700", marginBottom: "8px" }}>Hapus Permanen?</h3>
            <p style={{ color: "var(--text-muted)", fontSize: "14px", marginBottom: "20px" }}>
              Data <strong>{hardDeleteConfirm.nama || "tanpa nama"}</strong> akan dihapus permanen dari database dan <strong>tidak bisa dikembalikan</strong>.
            </p>
            <div style={{ display: "flex", justifyContent: "center", gap: "12px" }}>
              <button className="btn btn-secondary" onClick={() => setHardDeleteConfirm(null)} disabled={hardDeleteLoading}>Batal</button>
              <button className="btn btn-primary" style={{ background: "var(--danger)", borderColor: "var(--danger)" }} onClick={handleHardDeleteWarga} disabled={hardDeleteLoading}>
                {hardDeleteLoading ? "Menghapus..." : "Ya, Hapus Permanen"}
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}