import { create } from 'zustand';
import type { Session, SessionUser } from '@/domain/entities';

export type SessionStatus = 'loading' | 'authenticated' | 'anonymous';

interface SessionState {
  status: SessionStatus;
  accessToken: string | null;
  user: SessionUser | null;
  setSession: (session: Session) => void;
  clear: () => void;
}

/**
 * Estado de la sesión. El access token vive solo en memoria: nunca en
 * localStorage, para que un XSS no pueda robarlo de forma persistente.
 */
export const useSessionStore = create<SessionState>((set) => ({
  status: 'loading',
  accessToken: null,
  user: null,
  setSession: (session) => {
    set({ status: 'authenticated', accessToken: session.accessToken, user: session.user });
  },
  clear: () => {
    set({ status: 'anonymous', accessToken: null, user: null });
  },
}));
