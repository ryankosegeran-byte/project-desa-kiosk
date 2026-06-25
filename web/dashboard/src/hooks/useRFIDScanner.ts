import { useEffect, useRef, useState, useCallback } from "react";
import { API_BASE } from "../lib/api";

/**
 * RFID input for the admin panel. Listens to THREE sources simultaneously:
 *
 * 1. **Keyboard-wedge** (HID plug-n-play): rapid keystrokes + Enter on the
 *    operator's laptop. Works with no extra software.
 *
 * 2. **Server relay** (SSE from online server): the kiosk desa scans a card
 *    with its ACR122U, POSTs the UID to the server, and the server streams it
 *    here. Requires kiosk + server both online. This is the recommended flow
 *    for linking cards - ACR122U stays at the kiosk, no extra hardware needed
 *    on the PIC's laptop.
 *
 * 3. **Local bridge agent** (optional, rfid-agent.exe on the PIC's laptop):
 *    for PICs that have their own ACR122U reader. Connects via SSE to
 *    localhost:8088. Totally optional.
 *
 * The hook de-duplicates UIDs arriving from multiple sources within a short
 * window to avoid double-firing.
 */
export type KioskRelayStatus = "connecting" | "online" | "offline";
export type BridgeStatus = "connecting" | "online" | "offline" | "disabled";

const BRIDGE_URL =
  (import.meta.env.VITE_RFID_BRIDGE_URL as string | undefined) ??
  "http://localhost:8088";

export function useRFIDScanner(
  active: boolean,
  onScan: (uid: string) => void,
  desaId?: string,
): {
  kioskRelayStatus: KioskRelayStatus;
  bridgeStatus: BridgeStatus;
} {
  const onScanRef = useRef(onScan);
  onScanRef.current = onScan;

  const [kioskRelayStatus, setKioskRelayStatus] =
    useState<KioskRelayStatus>("connecting");
  const [bridgeStatus, setBridgeStatus] = useState<BridgeStatus>("disabled");

  // Dedup: ignore same UID within 1.5s across sources
  const lastUID = useRef("");
  const lastAt = useRef(0);

  const emit = useCallback((uid: string) => {
    const now = Date.now();
    if (uid === lastUID.current && now - lastAt.current < 1500) return;
    lastUID.current = uid;
    lastAt.current = now;
    onScanRef.current(uid);
  }, []);

  useEffect(() => {
    if (!active) {
      setKioskRelayStatus("connecting");
      setBridgeStatus("disabled");
      return;
    }

    const cleanups: (() => void)[] = [];

    // ----- Source 1: Server relay SSE (kiosk -> server -> admin panel) -----
    {
      setKioskRelayStatus("connecting");
      const token = localStorage.getItem("token");
      // EventSource does not support custom headers, so we pass the token as a
      // query parameter. The server's AuthMiddleware already reads from the
      // Authorization header, so we use a lightweight fetch-based SSE reader.
      const controller = new AbortController();
      let retryTimer: ReturnType<typeof setTimeout> | null = null;

      const connectServerSSE = () => {
        const streamUrl = desaId
          ? `${API_BASE}/api/rfid/stream?desa_id=${encodeURIComponent(desaId)}`
          : `${API_BASE}/api/rfid/stream`;
        fetch(streamUrl, {
          headers: { Authorization: `Bearer ${token}` },
          signal: controller.signal,
        })
          .then(async (res) => {
            if (!res.ok || !res.body) {
              setKioskRelayStatus("offline");
              retryTimer = setTimeout(connectServerSSE, 10000);
              return;
            }
            setKioskRelayStatus("online");
            const reader = res.body.getReader();
            const decoder = new TextDecoder();
            let buffer = "";

            // eslint-disable-next-line no-constant-condition
            while (true) {
              const { done, value } = await reader.read();
              if (done) break;
              buffer += decoder.decode(value, { stream: true });
              const lines = buffer.split("\n");
              buffer = lines.pop() ?? "";
              for (const line of lines) {
                if (line.startsWith("data: ")) {
                  const uid = line.slice(6).trim();
                  if (uid) emit(uid);
                }
              }
            }
            // Stream ended (server closed / network dropped)
            setKioskRelayStatus("offline");
            retryTimer = setTimeout(connectServerSSE, 5000);
          })
          .catch(() => {
            setKioskRelayStatus("offline");
            retryTimer = setTimeout(connectServerSSE, 10000);
          });
      };

      connectServerSSE();
      cleanups.push(() => {
        controller.abort();
        if (retryTimer) clearTimeout(retryTimer);
      });
    }

    // ----- Source 2: Local bridge agent SSE (optional ACR122U on laptop) -----
    {
      let es: EventSource | null = null;
      setBridgeStatus("connecting");
      try {
        es = new EventSource(`${BRIDGE_URL}/api/rfid/events`);
        es.onopen = () => setBridgeStatus("online");
        es.onmessage = (event) => {
          const uid = (event.data || "").trim();
          if (uid) {
            setBridgeStatus("online");
            emit(uid);
          }
        };
        es.onerror = () => setBridgeStatus("offline");
      } catch {
        setBridgeStatus("offline");
      }
      cleanups.push(() => es?.close());
    }

    // ----- Source 3: Keyboard-wedge (HID) -----
    {
      const buffer: string[] = [];
      let lastKeyTime = Date.now();

      const handleKeyDown = (e: KeyboardEvent) => {
        const now = Date.now();
        if (now - lastKeyTime > 50) buffer.length = 0;
        lastKeyTime = now;

        if (e.key === "Enter") {
          if (buffer.length >= 4) {
            const uid = buffer.join("").trim();
            buffer.length = 0;
            emit(uid);
          }
        } else if (e.key.length === 1) {
          buffer.push(e.key);
        }
      };

      window.addEventListener("keydown", handleKeyDown);
      cleanups.push(() => window.removeEventListener("keydown", handleKeyDown));
    }

    return () => cleanups.forEach((fn) => fn());
  }, [active, emit, desaId]);

  return { kioskRelayStatus, bridgeStatus };
}