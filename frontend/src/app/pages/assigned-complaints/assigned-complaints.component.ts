import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TableModule } from 'primeng/table';
import { TagModule } from 'primeng/tag';
import { Complaint } from '../../core/models/complaint.model';
import { ComplaintService } from '../../core/services/complaint.service';
import { AdminAuthService } from '../../core/services/admin-auth.service';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { finalize } from 'rxjs';

@Component({
  selector: 'app-assigned-complaints',
  standalone: true,
  imports: [CommonModule, TableModule, TagModule, NavbarComponent],
  templateUrl: './assigned-complaints.component.html',
  styleUrls: ['./assigned-complaints.component.scss']
})
export class AssignedComplaintsComponent implements OnInit {
  complaints: Complaint[] = [];
  solvedComplaints: Complaint[] = [];
  groupedComplaints: Record<string, Complaint[]> = {};
  groupedSolvedComplaints: Record<string, Complaint[]> = {};
  visibleTeams: string[] = [];
  visibleSolvedTeams: string[] = [];
  currentAdminRole = '';
  currentAdminScope = '';
  currentAdminTeam = '';
  isLoading = false;
  solvingComplaintId = '';

  constructor(
    private complaintService: ComplaintService,
    private authService: AdminAuthService
  ) {}

  ngOnInit(): void {
    const currentAdmin = this.authService.getCurrentAdmin();
    this.currentAdminRole = currentAdmin?.role ?? '';
    this.currentAdminScope = currentAdmin?.app_scope ?? '';
    this.currentAdminTeam = this.getExpectedTeamForCurrentAdmin();
    this.loadAssignedComplaints();
  }

  loadAssignedComplaints(): void {
    this.isLoading = true;
    this.complaintService.loadComplaints(1, 200, '').subscribe({
      next: (response) => {
        const teamFiltered = response.data.filter((complaint) => {
          const team = this.normalizeRole(complaint.escalation?.current_team ?? '');
          return !!team && team === this.currentAdminTeam;
        });
        this.complaints = teamFiltered.filter((complaint) => this.isOpenComplaint(complaint));
        this.solvedComplaints = teamFiltered.filter((complaint) => !this.isOpenComplaint(complaint));
        this.buildViewModel();
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Failed to load assigned complaints:', err);
        this.isLoading = false;
      }
    });
  }

  getRoleLabel(team: string): string {
    const key = this.normalizeRole(team);
    if (key === 'driver_admin') return 'Driver Admin';
    if (key === 'driver_supervisor') return 'Driver Supervisor';
    if (key === 'super_admin') return 'Super Admin';
    return team.replace(/_/g, ' ').replace(/\b\w/g, ch => ch.toUpperCase());
  }

  formatDate(dateString?: string): string {
    if (!dateString) return 'No deadline';
    return new Date(dateString).toLocaleString();
  }

  getTimeRemaining(nextDue?: string): string {
    if (!nextDue) return 'No SLA deadline';

    const due = new Date(nextDue).getTime();
    const now = Date.now();
    const diffMs = due - now;
    const abs = Math.abs(diffMs);
    const days = Math.floor(abs / (1000 * 60 * 60 * 24));
    const hours = Math.floor((abs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));

    if (diffMs >= 0) {
      return `${days}d ${hours}h left`;
    }

    return `Overdue by ${days}d ${hours}h`;
  }

  getDeadlineSeverity(nextDue?: string): 'success' | 'warn' | 'danger' | 'secondary' {
    if (!nextDue) return 'secondary';

    const due = new Date(nextDue).getTime();
    const now = Date.now();
    const diff = due - now;

    if (diff < 0) return 'danger';
    if (diff <= 12 * 60 * 60 * 1000) return 'warn';
    return 'success';
  }

  canSolveComplaint(complaint: Complaint): boolean {
    const status = this.normalizeRole(complaint.status);
    return status !== 'resolved' && status !== 'closed';
  }

  solveComplaint(complaint: Complaint): void {
    if (!complaint?.id || !this.canSolveComplaint(complaint) || this.solvingComplaintId) {
      return;
    }

    const resolutionNote = window.prompt('Enter solution note for this complaint:');
    if (resolutionNote === null) {
      return;
    }

    const trimmedNote = resolutionNote.trim();
    if (!trimmedNote) {
      alert('Solution note is required to solve the complaint.');
      return;
    }

    this.solvingComplaintId = complaint.id;
    this.complaintService.updateComplaintStatus(complaint.id, 'resolved', trimmedNote)
      .pipe(finalize(() => {
        this.solvingComplaintId = '';
      }))
      .subscribe({
      next: () => {
        alert('Complaint solved successfully.');
        this.loadAssignedComplaints();
      },
      error: (err) => {
        console.error('Failed to solve complaint:', err);
        const backendMessage = err?.error?.error || err?.error?.message || err?.message || 'Please try again.';
        alert(`Failed to solve complaint: ${backendMessage}`);
      }
    });
  }

  private buildViewModel(): void {
    const grouped: Record<string, Complaint[]> = {};
    const groupedSolved: Record<string, Complaint[]> = {};

    this.complaints.forEach((complaint) => {
      const team = complaint.escalation?.current_team;
      if (!team) return;
      if (!grouped[team]) {
        grouped[team] = [];
      }
      grouped[team].push(complaint);
    });

    this.solvedComplaints.forEach((complaint) => {
      const team = complaint.escalation?.current_team;
      if (!team) return;
      if (!groupedSolved[team]) {
        groupedSolved[team] = [];
      }
      groupedSolved[team].push(complaint);
    });

    this.groupedComplaints = grouped;
    this.groupedSolvedComplaints = groupedSolved;
    this.visibleTeams = this.getVisibleTeams(grouped);
    this.visibleSolvedTeams = this.getVisibleTeams(groupedSolved);
  }

  private getVisibleTeams(grouped: Record<string, Complaint[]>): string[] {
    return Object.keys(grouped).filter((team) => this.normalizeRole(team) === this.currentAdminTeam);
  }

  getComplaintsForTeam(team: string): Complaint[] {
    return this.groupedComplaints[team] ?? [];
  }

  getSolvedComplaintsForTeam(team: string): Complaint[] {
    return this.groupedSolvedComplaints[team] ?? [];
  }

  private getExpectedTeamForCurrentAdmin(): string {
    const role = this.normalizeRole(this.currentAdminRole);
    const scope = this.normalizeRole(this.currentAdminScope);

    if (role === 'super_admin') {
      return 'super_admin';
    }

    if (!scope || (role !== 'admin' && role !== 'supervisor')) {
      return '';
    }

    return `${scope}_${role}`;
  }

  private isSuperAdmin(): boolean {
    return this.normalizeRole(this.currentAdminRole) === 'super_admin';
  }

  private isOpenComplaint(complaint: Complaint): boolean {
    const status = this.normalizeRole(complaint.status);
    return status !== 'resolved' && status !== 'closed';
  }

  private normalizeRole(role: string): string {
    return (role || '').toLowerCase().trim().replace(/[-\s]+/g, '_');
  }
}
