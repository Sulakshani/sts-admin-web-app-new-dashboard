export interface Complaint {
  id: string;
  role: string;
  name: string;
  contact: string;
  category: string;
  message: string;
  media: string;
  dateTime: string;
  assignedTeam: string;
  resolvedBy: string;
  activity: string;
  status: string;
  priority: string;
  latitude?: number;
  longitude?: number;
  locationAddress?: string;
  escalation?: ComplaintEscalation;
}

export interface ComplaintEscalation {
  current_level: number;
  current_team: string;
  source_app?: string;
  assigned_to_admin_id?: string;
  previous_assigned_admin_id?: string;
  assigned_to_name?: string;
  last_escalated_at?: string;
  next_escalation_due?: string;
  escalation_history?: EscalationHistoryEntry[];
}

export interface EscalationHistoryEntry {
  level: number;
  team_name: string;
  escalated_at: string;
  escalated_by: string;
  from_admin_id?: string;
  to_admin_id?: string;
  reason: string;
}

export interface PaginatedComplaintsResponse {
  data: Complaint[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}
