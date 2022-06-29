import { getLeader } from '@/nomad/rest/getLeader';

// Use getLeader as ping to check nomad aliveness
export const ping = getLeader;
