import { useEffect, useRef } from 'react';

/**
 * useRFID hook listens to both:
 * 1. SSE events stream from Go backend (good for mock scans or advanced serial readers)
 * 2. Rapid keystrokes on window (good for standard plug-and-play USB Keyboard Wedge readers)
 */
export function useRFID(onScan: (uid: string) => void) {
  const onScanRef = useRef(onScan);
  onScanRef.current = onScan;

  useEffect(() => {
    // ==========================================
    // Method 1: SSE Events from Go Backend
    // ==========================================
    const apiBase = import.meta.env.DEV ? 'http://localhost:8080' : '';
    const eventSource = new EventSource(`${apiBase}/api/rfid/events`);

    eventSource.onmessage = (event) => {
      if (event.data) {
        console.log("RFID scan received via SSE:", event.data);
        onScanRef.current(event.data);
      }
    };

    eventSource.onerror = (error) => {
      console.warn("RFID EventSource error (expected if local backend is down):", error);
    };

    // ==========================================
    // Method 2: Keyboard Wedge HID Reader
    // ==========================================
    let buffer = '';
    let lastKeyTime = Date.now();

    const handleKeyDown = (e: KeyboardEvent) => {
      const currentTime = Date.now();
      const timeDiff = currentTime - lastKeyTime;
      lastKeyTime = currentTime;

      // Most RFID readers send characters rapidly (< 50ms apart) and terminate with 'Enter'
      if (timeDiff > 50) {
        // Clear buffer if user is typing manually (slow delay)
        buffer = '';
      }

      // Ignore modifier keys
      if (e.key.length === 1) {
        buffer += e.key;
      } else if (e.key === 'Enter') {
        if (buffer.length >= 4) { // RFID UIDs are typically 4+ characters
          e.preventDefault();
          const scanUID = buffer.trim();
          buffer = '';
          console.log("RFID scan received via Keyboard Wedge:", scanUID);
          onScanRef.current(scanUID);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      eventSource.close();
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, []);
}
