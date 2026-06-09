import React, { useState, useEffect, useCallback } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useOnlineStatus } from './hooks/useOnlineStatus';
import { StatusBar } from './components/StatusBar';

// Pages
import { HomePage } from './pages/HomePage';
import { SelectSuratPage } from './pages/SelectSuratPage';
import { FormSuratPage } from './pages/FormSuratPage';
import { PreviewPage } from './pages/PreviewPage';
import { SuccessPage } from './pages/SuccessPage';

// Types
import type { Warga } from './hooks/useWarga';
import type { JenisSurat } from './hooks/useSurat';

export const AppContent: React.FC = () => {
  const [warga, setWarga] = useState<Warga | null>(null);
  const [jenisSurat, setJenisSurat] = useState<JenisSurat | null>(null);
  const [formData, setFormData] = useState<any>({});
  const { kioskName } = useOnlineStatus();

  // Reset all flow states
  const handleReset = useCallback(() => {
    setWarga(null);
    setJenisSurat(null);
    setFormData({});
  }, []);

  // Idle timer auto-reset: Reset and send to home screen if idle for 60 seconds
  useEffect(() => {
    // We only track idle timer if a user is logged in / identified
    if (!warga) return;

    let idleTimeout: number;

    const resetIdleTimer = () => {
      clearTimeout(idleTimeout);
      idleTimeout = window.setTimeout(() => {
        console.log("Kiosk idle timeout reached. Resetting flow to home page.");
        handleReset();
        window.location.href = '/'; // Reset browser state and route to home
      }, 60000); // 60 seconds
    };

    // Listen for touch/mouse/keyboard activity
    const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
    
    activityEvents.forEach(event => {
      document.addEventListener(event, resetIdleTimer);
    });

    // Start timer on initial mount
    resetIdleTimer();

    return () => {
      clearTimeout(idleTimeout);
      activityEvents.forEach(event => {
        document.removeEventListener(event, resetIdleTimer);
      });
    };
  }, [warga, handleReset]);

  return (
    <div className="kiosk-container">
      {/* Brand Header */}
      <header className="kiosk-header">
        <div className="kiosk-logo-area">
          <div className="kiosk-logo-icon">K</div>
          <div className="kiosk-title">
            <h1>PELAYANAN MANDIRI</h1>
            <p>Kantor {kioskName}</p>
          </div>
        </div>
        
        {warga && (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', fontSize: '14px' }}>
            <span style={{ color: 'var(--text-muted)' }}>Warga Aktif</span>
            <span style={{ fontWeight: 600 }}>{warga.nama}</span>
          </div>
        )}
      </header>

      {/* Main Pages Viewport */}
      <main style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', marginBottom: '24px' }}>
        <Routes>
          <Route
            path="/"
            element={<HomePage onWargaIdentified={setWarga} />}
          />
          <Route
            path="/select-surat"
            element={
              warga ? (
                <SelectSuratPage
                  warga={warga}
                  onSuratSelected={setJenisSurat}
                  onCancel={handleReset}
                />
              ) : (
                <Navigate to="/" replace />
              )
            }
          />
          <Route
            path="/form-surat"
            element={
              warga && jenisSurat ? (
                <FormSuratPage
                  warga={warga}
                  jenisSurat={jenisSurat}
                  onFormSubmitted={setFormData}
                />
              ) : (
                <Navigate to="/" replace />
              )
            }
          />
          <Route
            path="/preview"
            element={
              warga && jenisSurat ? (
                <PreviewPage
                  warga={warga}
                  jenisSurat={jenisSurat}
                  formData={formData}
                  onSuratPrinted={handleReset}
                />
              ) : (
                <Navigate to="/" replace />
              )
            }
          />
          <Route
            path="/success"
            element={<SuccessPage onReset={handleReset} />}
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>

      {/* Always visible Footer Status Bar */}
      <StatusBar />
    </div>
  );
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
};

export default App;
