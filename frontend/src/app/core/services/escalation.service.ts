import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';

export interface EscalationLevel {
  level: number;
  team_name: string;
  assigned_to: string[];
  escalation_days: number;
  notification_emails?: string[];
}

export interface EscalationConfig {
  category: string;
  levels: EscalationLevel[];
}

export interface EscalationHistoryEntry {
  level: number;
  team_name: string;
  escalated_at: string;
  escalated_by: string;
  reason: string;
}

export interface ComplaintEscalation {
  current_level: number;
  current_team: string;
  assigned_to_admin_id?: string;
  assigned_to_name?: string;
  last_escalated_at?: string;
  next_escalation_due?: string;
  escalation_history?: EscalationHistoryEntry[];
}

export interface EscalationStats {
  by_level: { [key: number]: number };
  overdue_count: number;
}

@Injectable({
  providedIn: 'root'
})
export class EscalationService {
  private apiUrl = `${environment.apiUrl}/escalation`;

  constructor(private http: HttpClient) {}

  /**
   * Get escalation configuration for all categories
   */
  getAllConfigs(): Observable<{ configs: { [key: string]: EscalationConfig } }> {
    return this.http.get<{ configs: { [key: string]: EscalationConfig } }>(
      `${this.apiUrl}/config`
    );
  }

  /**
   * Get escalation configuration for a specific category
   */
  getConfigForCategory(category: string): Observable<EscalationConfig> {
    return this.http.get<EscalationConfig>(
      `${this.apiUrl}/config/${category}`
    );
  }

  /**
   * Get escalation status for a specific complaint
   */
  getComplaintEscalation(complaintId: string): Observable<ComplaintEscalation> {
    return this.http.get<ComplaintEscalation>(
      `${this.apiUrl}/complaint/${complaintId}`
    );
  }

  /**
   * Get escalation history for a complaint
   */
  getEscalationHistory(complaintId: string): Observable<{ history: EscalationHistoryEntry[] }> {
    return this.http.get<{ history: EscalationHistoryEntry[] }>(
      `${this.apiUrl}/complaint/${complaintId}/history`
    );
  }

  /**
   * Manually escalate a complaint to the next level
   */
  escalateComplaint(
    complaintId: string,
    category: string,
    currentLevel: number,
    escalatedBy?: string
  ): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(
      `${this.apiUrl}/complaint/${complaintId}/escalate`,
      {
        category,
        current_level: currentLevel,
        escalated_by: escalatedBy || 'admin'
      }
    );
  }

  /**
   * Assign a complaint to a specific admin
   */
  assignComplaint(
    complaintId: string,
    adminId: string
  ): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(
      `${this.apiUrl}/complaint/${complaintId}/assign`,
      { admin_id: adminId }
    );
  }

  /**
   * Initialize escalation for a new complaint
   */
  initializeEscalation(
    complaintId: string,
    category: string
  ): Observable<{ message: string }> {
    return this.http.post<{ message: string }>(
      `${this.apiUrl}/complaint/${complaintId}/initialize`,
      { category }
    );
  }

  /**
   * Get escalation statistics
   */
  getEscalationStats(): Observable<EscalationStats> {
    return this.http.get<EscalationStats>(`${this.apiUrl}/stats`);
  }

  /**
   * Calculate days until next escalation
   */
  getDaysUntilEscalation(nextEscalationDue: string): number {
    const due = new Date(nextEscalationDue);
    const now = new Date();
    const diffTime = due.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  }

  /**
   * Get badge class for escalation level
   */
  getLevelBadgeClass(level: number): string {
    switch (level) {
      case 1:
        return 'badge-info';
      case 2:
        return 'badge-warning';
      case 3:
        return 'badge-danger';
      default:
        return 'badge-secondary';
    }
  }

  /**
   * Check if complaint is overdue for escalation
   */
  isOverdue(nextEscalationDue: string): boolean {
    const due = new Date(nextEscalationDue);
    const now = new Date();
    return due < now;
  }
}
