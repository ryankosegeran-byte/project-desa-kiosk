import { useEffect, useState } from 'react';

export interface RegistrationSession {
  active: boolean;
  name: string;
}

const API_BASE = import.meta.env.DEV ? 'http://localhost:8080' : '';

/**
 * useRegistrationSession polls the kiosk backend for whether a registration
 * session is active (driven by an operator in the admin panel via the server),
 * and the name of the warga being registered.
 */
export function useRegistrationSession(pollMs = 2000): RegistrationSession {
  const [session, setSession] = useState<RegistrationSession>({ active: false, name: '' });

  useEffect(() => {
    let cancelled = false;

    const poll = async () => {
      try {
        const res = await fetch(`${API_BASE}/api/rfid/session`);
        if (!res.ok) return;
        const data = await res.json();
        if (!cancelled) setSession({ active: Boolean(data.active), name: data.name || '' });
      } catch {
        if (!cancelled) setSession({ active: false, name: '' });
      }
    };

    poll();
    const id = setInterval(poll, pollMs);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [pollMs]);

  return session;
}

/**
 * reportKioskBusy tells the kiosk backend (and via it, the server/admin panel)
 * whether a resident is currently mid-flow creating a letter.
 */
export async function reportKioskBusy(busy: boolean): Promise<void> {
  try {
    await fetch(`${API_BASE}/api/rfid/busy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ busy }),
    });
  } catch {
    // best-effort
  }
}