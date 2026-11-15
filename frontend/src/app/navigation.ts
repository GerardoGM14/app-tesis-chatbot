import {
  Bot,
  Building2,
  LayoutDashboard,
  type LucideIcon,
  MessagesSquare,
  QrCode,
  Users,
} from 'lucide-react';
import type { Permission } from '@/domain/entities';

export interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
  permission: Permission;
}

/** Ítems del sidebar; cada usuario ve solo los que su rol permite. */
export const NAV_ITEMS: NavItem[] = [
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, permission: 'dashboard:ver' },
  {
    to: '/conversaciones',
    label: 'Conversaciones',
    icon: MessagesSquare,
    permission: 'conversaciones:atender',
  },
  { to: '/bot', label: 'Bot', icon: Bot, permission: 'bot:gestionar' },
  { to: '/canal', label: 'Canal de WhatsApp', icon: QrCode, permission: 'canal:gestionar' },
  { to: '/usuarios', label: 'Usuarios', icon: Users, permission: 'usuarios:gestionar' },
  { to: '/empresas', label: 'Empresas', icon: Building2, permission: 'empresas:gestionar' },
];

export function visibleNavItems(permissions: readonly Permission[]): NavItem[] {
  return NAV_ITEMS.filter((item) => permissions.includes(item.permission));
}
