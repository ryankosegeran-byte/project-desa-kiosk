import React, { useEffect, useState } from "react";
import { request } from "../lib/api";

interface User {
  id: string;
  username: string;
  nama: string;
  role: string;
  jabatan?: string;
  desa_id?: string;
  active: boolean;
  last_login_at?: string;
}

interface Desa {
  id: string;
  nama: string;
}

export default function UserManager() {
  const [users, setUsers] = useState<User[]>([]);
  const [desas, setDesas] = useState<Desa[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Modals state
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showPwdModal, setShowPwdModal] = useState(false);
  
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  // Form states
  const [addForm, setAddForm] = useState({
    username: "",
    password: "",
    nama: "",
    role: "pic_desa",
    jabatan: "",
    desa_id: "",
  });

  const [editForm, setEditForm] = useState({
    username: "",
    nama: "",
    role: "pic_desa",
    jabatan: "",
    desa_id: "",
    active: true,
  });

  const [newPassword, setNewPassword] = useState("");

  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      const userData = await request("/api/users");
      setUsers(userData);

      const desaData = await request("/api/desa");
      setDesas(desaData);
      if (desaData.length > 0) {
        setAddForm((prev) => ({ ...prev, desa_id: desaData[0].id }));
      }
    } catch (err: any) {
      setError(err.message || "Gagal mengambil data user.");
    } finally {
      setLoading(false);
    }
  }

  const handleAddSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setActionError("");

    try {
      const payload = {
        ...addForm,
        desa_id: addForm.role === "superadmin" ? undefined : addForm.desa_id,
      };

      const created = await request("/api/users", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      setUsers([...users, created]);
      setShowAddModal(false);
      setAddForm({
        username: "",
        password: "",
        nama: "",
        role: "pic_desa",
        jabatan: "",
        desa_id: desas[0]?.id || "",
      });
    } catch (err: any) {
      setActionError(err.message || "Gagal membuat user.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleEditClick = (u: User) => {
    setSelectedUser(u);
    setEditForm({
      username: u.username,
      nama: u.nama,
      role: u.role,
      jabatan: u.jabatan || "",
      desa_id: u.desa_id || (desas[0]?.id || ""),
      active: u.active,
    });
    setShowEditModal(true);
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedUser) return;
    setActionLoading(true);
    setActionError("");

    try {
      const payload = {
        ...editForm,
        desa_id: editForm.role === "superadmin" ? undefined : editForm.desa_id,
      };

      const updated = await request(`/api/users/${selectedUser.id}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });

      setUsers(users.map((u) => (u.id === selectedUser.id ? updated : u)));
      setShowEditModal(false);
      setSelectedUser(null);
    } catch (err: any) {
      setActionError(err.message || "Gagal memperbarui user.");
    } finally {
      setActionLoading(false);
    }
  };

  const handleResetPwdSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedUser) return;
    setActionLoading(true);
    setActionError("");

    try {
      await request(`/api/users/${selectedUser.id}/reset-password`, {
        method: "PUT",
        body: JSON.stringify({ new_password: newPassword }),
      });
      setShowPwdModal(false);
      setNewPassword("");
      setSelectedUser(null);
      alert("Password berhasil direset!");
    } catch (err: any) {
      setActionError(err.message || "Gagal meriset password.");
    } finally {
      setActionLoading(false);
    }
  };

  const getDesaName = (desaId?: string) => {
    if (!desaId) return "-";
    const d = desas.find((x) => x.id === desaId);
    return d ? d.nama : "Tidak Ditemukan";
  };

  if (loading) {
    return (
      <div style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "400px" }}>
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "32px" }}>
        <div>
          <h1 style={{ fontSize: "28px", fontWeight: "700" }}>Kelola User PIC & Kiosk</h1>
          <p style={{ color: "var(--text-muted)", marginTop: "4px" }}>
            Kelola hak akses untuk petugas/perangkat desa (PIC Desa) dan terminal kiosk.
          </p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowAddModal(true)}>
          ➕ Daftarkan User Baru
        </button>
      </div>

      {error && (
        <div className="glass-card" style={{ borderLeft: "4px solid var(--danger)", color: "var(--danger)", marginBottom: "24px" }}>
          ⚠️ {error}
        </div>
      )}

      {/* Users Table */}
      <div className="table-container">
        <table className="premium-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>Nama Lengkap</th>
              <th>Role</th>
              <th>Jabatan</th>
              <th>Desa Tugas</th>
              <th>Status</th>
              <th>Login Terakhir</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id}>
                <td style={{ fontWeight: "600" }}>{u.username}</td>
                <td>{u.nama}</td>
                <td>
                  <span className={`badge ${u.role === "superadmin" ? "badge-danger" : u.role === "pic_desa" ? "badge-primary" : "badge-success"}`}>
                    {u.role}
                  </span>
                </td>
                <td>{u.jabatan || "-"}</td>
                <td>{getDesaName(u.desa_id)}</td>
                <td>
                  {u.active ? (
                    <span className="badge badge-success">Aktif</span>
                  ) : (
                    <span className="badge badge-danger">Nonaktif</span>
                  )}
                </td>
                <td>
                  {u.last_login_at
                    ? new Date(u.last_login_at).toLocaleString("id-ID", { dateStyle: "short", timeStyle: "short" })
                    : "-"}
                </td>
                <td>
                  <div style={{ display: "flex", gap: "8px" }}>
                    <button className="btn btn-secondary" style={{ padding: "6px 12px", fontSize: "13px" }} onClick={() => handleEditClick(u)}>
                      ✏️ Edit
                    </button>
                    <button
                      className="btn btn-secondary"
                      style={{ padding: "6px 12px", fontSize: "13px" }}
                      onClick={() => {
                        setSelectedUser(u);
                        setShowPwdModal(true);
                      }}
                    >
                      🔑 Reset Password
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Add User Modal */}
      {showAddModal && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "rgba(0,0,0,0.7)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "500px", width: "95%", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px" }}>Daftarkan User Baru</h3>

            {actionError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {actionError}
              </div>
            )}

            <form onSubmit={handleAddSubmit}>
              <div className="form-group">
                <label className="form-label">Username</label>
                <input
                  type="text"
                  className="form-control"
                  value={addForm.username}
                  onChange={(e) => setAddForm({ ...addForm, username: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Password</label>
                <input
                  type="password"
                  className="form-control"
                  value={addForm.password}
                  onChange={(e) => setAddForm({ ...addForm, password: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Nama Lengkap</label>
                <input
                  type="text"
                  className="form-control"
                  value={addForm.nama}
                  onChange={(e) => setAddForm({ ...addForm, nama: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Role</label>
                <select
                  className="form-control"
                  value={addForm.role}
                  onChange={(e) => setAddForm({ ...addForm, role: e.target.value })}
                  disabled={actionLoading}
                >
                  <option value="pic_desa">PIC Perangkat Desa</option>
                  <option value="superadmin">Superadmin Hub</option>
                  <option value="kiosk">Sistem Kiosk (Sync)</option>
                </select>
              </div>

              {addForm.role !== "superadmin" && (
                <>
                  <div className="form-group">
                    <label className="form-label">Jabatan (Opsional)</label>
                    <input
                      type="text"
                      className="form-control"
                      placeholder="Contoh: Sekretaris Desa"
                      value={addForm.jabatan}
                      onChange={(e) => setAddForm({ ...addForm, jabatan: e.target.value })}
                      disabled={actionLoading}
                    />
                  </div>

                  {desas.length > 0 && (
                    <div className="form-group">
                      <label className="form-label">Tugas di Desa</label>
                      <select
                        className="form-control"
                        value={addForm.desa_id}
                        onChange={(e) => setAddForm({ ...addForm, desa_id: e.target.value })}
                        disabled={actionLoading}
                      >
                        {desas.map((d) => (
                          <option key={d.id} value={d.id}>
                            {d.nama}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}
                </>
              )}

              <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowAddModal(false)} disabled={actionLoading}>
                  Batal
                </button>
                <button type="submit" className="btn btn-primary" disabled={actionLoading}>
                  {actionLoading ? "Menyimpan..." : "Daftarkan"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit User Modal */}
      {showEditModal && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "rgba(0,0,0,0.7)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "500px", width: "95%", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "20px" }}>Edit User</h3>

            {actionError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {actionError}
              </div>
            )}

            <form onSubmit={handleEditSubmit}>
              <div className="form-group">
                <label className="form-label">Username</label>
                <input
                  type="text"
                  className="form-control"
                  value={editForm.username}
                  onChange={(e) => setEditForm({ ...editForm, username: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Nama Lengkap</label>
                <input
                  type="text"
                  className="form-control"
                  value={editForm.nama}
                  onChange={(e) => setEditForm({ ...editForm, nama: e.target.value })}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div className="form-group">
                <label className="form-label">Role</label>
                <select
                  className="form-control"
                  value={editForm.role}
                  onChange={(e) => setEditForm({ ...editForm, role: e.target.value })}
                  disabled={actionLoading}
                >
                  <option value="pic_desa">PIC Perangkat Desa</option>
                  <option value="superadmin">Superadmin Hub</option>
                  <option value="kiosk">Sistem Kiosk (Sync)</option>
                </select>
              </div>

              {editForm.role !== "superadmin" && (
                <>
                  <div className="form-group">
                    <label className="form-label">Jabatan (Opsional)</label>
                    <input
                      type="text"
                      className="form-control"
                      value={editForm.jabatan}
                      onChange={(e) => setEditForm({ ...editForm, jabatan: e.target.value })}
                      disabled={actionLoading}
                    />
                  </div>

                  {desas.length > 0 && (
                    <div className="form-group">
                      <label className="form-label">Tugas di Desa</label>
                      <select
                        className="form-control"
                        value={editForm.desa_id}
                        onChange={(e) => setEditForm({ ...editForm, desa_id: e.target.value })}
                        disabled={actionLoading}
                      >
                        {desas.map((d) => (
                          <option key={d.id} value={d.id}>
                            {d.nama}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}
                </>
              )}

              <div className="form-group" style={{ flexDirection: "row", alignItems: "center", gap: "10px", marginTop: "10px" }}>
                <input
                  type="checkbox"
                  id="user-active-chk"
                  checked={editForm.active}
                  onChange={(e) => setEditForm({ ...editForm, active: e.target.checked })}
                  disabled={actionLoading}
                />
                <label htmlFor="user-active-chk" className="form-label" style={{ marginBottom: 0 }}>User Aktif</label>
              </div>

              <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowEditModal(false)} disabled={actionLoading}>
                  Batal
                </button>
                <button type="submit" className="btn btn-primary" disabled={actionLoading}>
                  {actionLoading ? "Menyimpan..." : "Simpan Perubahan"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Password Reset Modal */}
      {showPwdModal && (
        <div style={{ position: "fixed", top: 0, left: 0, right: 0, bottom: 0, background: "rgba(0,0,0,0.7)", backdropFilter: "blur(4px)", display: "flex", justifyContent: "center", alignItems: "center", zIndex: 1000 }}>
          <div className="glass-card" style={{ maxWidth: "400px", width: "95%", padding: "30px" }}>
            <h3 style={{ fontSize: "20px", fontWeight: "700", marginBottom: "16px" }}>Reset Password User</h3>
            <p style={{ fontSize: "14px", color: "var(--text-muted)", marginBottom: "20px" }}>
              Masukkan password baru untuk user: <strong>{selectedUser?.username}</strong>
            </p>

            {actionError && (
              <div style={{ color: "var(--danger)", background: "hsla(355, 85%, 55%, 0.1)", padding: "10px", borderRadius: "var(--radius-sm)", marginBottom: "16px", fontSize: "13px" }}>
                ⚠️ {actionError}
              </div>
            )}

            <form onSubmit={handleResetPwdSubmit}>
              <div className="form-group">
                <label className="form-label">Password Baru (min 6 karakter)</label>
                <input
                  type="password"
                  className="form-control"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  required
                  disabled={actionLoading}
                />
              </div>

              <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", marginTop: "24px" }}>
                <button type="button" className="btn btn-secondary" onClick={() => { setShowPwdModal(false); setNewPassword(""); }} disabled={actionLoading}>
                  Batal
                </button>
                <button type="submit" className="btn btn-primary" disabled={actionLoading}>
                  {actionLoading ? "Mengubah..." : "Ubah Password"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
