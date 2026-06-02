import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { TagModule } from 'primeng/tag';
import { DialogModule } from 'primeng/dialog';
import { TextareaModule } from 'primeng/textarea';
import { SelectModule } from 'primeng/select';
import { DatePickerModule } from 'primeng/datepicker';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NotificationService } from '../../core/services/notification.service';
import { ComplaintService } from '../../core/services/complaint.service';
import { Complaint } from '../../core/models/complaint.model';
import { AdminAuthService } from '../../core/services/admin-auth.service';
import { forkJoin, Observable } from 'rxjs';

@Component({
  selector: 'app-complaint-management',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    InputTextModule,
    TagModule,
    DialogModule,
    TextareaModule,
    SelectModule,
    DatePickerModule,
    NavbarComponent,
    NotificationPanelComponent,
    RouterModule
  ],
  templateUrl: './complaint-management.component.html',
  styleUrls: ['./complaint-management.component.scss']
})
export class ComplaintManagementComponent implements OnInit {
  activeTab: string = 'All';
  showNotificationPanel = false;
  showProfileMenu = false;
  
  stats = [
    { title: 'Total Complaints', count: 0, icon: 'pi pi-users', color: 'blue' },
    { title: 'Pending Complaints', count: 0, icon: 'pi pi-clock', color: 'indigo' },
    { title: 'Inprogress Complaints', count: 0, icon: 'pi pi-check-circle', color: 'purple' },
    { title: 'Resolved Complaints', count: 0, icon: 'pi pi-check', color: 'green' }
  ];

  complaints: Complaint[] = [];
  filteredComplaints: Complaint[] = [];
  page = 1;
  pageSize = 20;
  totalRecords = 0;
  totalPages = 0;
  isLoading = false;
  readonly pageSizeOptions = [10, 20, 50, 100];

  // Filter variables
  selectedCategory: string = '';
  selectedStatus: string = '';
  selectedAssignedTeam: string = '';
  selectedLevel: number | null = null;
  selectedDate: Date | null = null;

  // Filter options
  categoryOptions: any[] = [
    { label: 'All Categories', value: '' },
    { label: 'Bus Delay', value: 'Bus Delay' },
    { label: 'Maintenance Issue', value: 'Maintenance Issue' },
    { label: 'Flat Wheel', value: 'Flat Wheel' },
    { label: 'Passenger Complaint', value: 'Passenger Complaint' },
    { label: 'Safety Concern', value: 'Safety Concern' },
    { label: 'Equipment Malfunction', value: 'Equipment Malfunction' },
    { label: 'Other', value: 'Other' }
  ];

  statusOptions: any[] = [
    { label: 'All Statuses', value: '' },
    { label: 'Pending', value: 'pending' },
    { label: 'Acknowledged', value: 'acknowledged' },
    { label: 'In Progress', value: 'in progress' },
    { label: 'Resolved', value: 'resolved' },
    { label: 'Closed', value: 'closed' }
  ];

  assignedTeamOptions: any[] = [
    { label: 'All Teams', value: '' },
    { label: 'Level 1: Support Team', value: 'Level 1: Support Team' },
    { label: 'Level 2: Technical Team', value: 'Level 2: Technical Team' },
    { label: 'Level 3: Management Team', value: 'Level 3: Management Team' }
  ];

  levelOptions: any[] = [
    { label: 'All Levels', value: null },
    { label: 'Level 1', value: 1 },
    { label: 'Level 2', value: 2 },
    { label: 'Level 3', value: 3 }
  ];

  displayViewModal: boolean = false;
  displayEditModal: boolean = false;
  displayEscalationModal: boolean = false;
  selectedComplaint: any = {};
  originalAssignedTeam = '';
  canEscalateComplaint = false;
  currentAdminId = '';
  currentAdminRole = '';
  currentAdminScope = '';
  assignableRoles: Array<{ label: string; value: string }> = [];
  selectedAssigneeRole = '';
  isAssigningComplaint = false;

  get passengerComplaints(): Complaint[] {
    return this.activeTab === 'All'
      ? this.filteredComplaints.filter(c => c.role === 'Passenger')
      : this.filteredComplaints;
  }

  get driverComplaints(): Complaint[] {
    return this.activeTab === 'All'
      ? this.filteredComplaints.filter(c => c.role === 'Driver')
      : this.filteredComplaints;
  }

  get conductorComplaints(): Complaint[] {
    return this.activeTab === 'All'
      ? this.filteredComplaints.filter(c => c.role === 'Conductor')
      : this.filteredComplaints;
  }

  get busOwnerComplaints(): Complaint[] {
    return this.activeTab === 'All'
      ? this.filteredComplaints.filter(c => c.role === 'Bus Owner')
      : this.filteredComplaints;
  }

  get loungeOwnerComplaints(): Complaint[] {
    return this.activeTab === 'All'
      ? this.filteredComplaints.filter(c => c.role === 'Lounge Owner')
      : this.filteredComplaints;
  }
  solutionText: string = '';
  escalationInfo: any = null;
  isEscalating: boolean = false;

  constructor(
    private router: Router, 
    public notificationService: NotificationService,
    private complaintService: ComplaintService,
    private authService: AdminAuthService
  ) {}

  ngOnInit() {
    const currentAdmin = this.authService.getCurrentAdmin();
    this.currentAdminId = currentAdmin?.id ?? '';
    this.currentAdminRole = currentAdmin?.role ?? '';
    this.currentAdminScope = currentAdmin?.app_scope ?? '';
    this.canEscalateComplaint = this.isSuperAdmin() || this.normalizeRole(this.currentAdminRole) === 'supervisor';
    if (this.isSuperAdmin()) {
      this.setAssignableRoles();
    }
    this.loadComplaints(1);
  }

  canUpdateComplaint(complaint: Complaint): boolean {
    if (this.isSuperAdmin()) {
      return true;
    }

    const expectedTeam = this.getExpectedTeamForCurrentAdmin();
    return expectedTeam !== '' && complaint?.escalation?.current_team === expectedTeam;
  }

  viewComplaint(complaint: any) {
    this.selectedComplaint = { ...complaint };
    this.solutionText = ''; 
    this.selectedAssigneeRole = complaint?.escalation?.current_team ?? '';
    this.loadEscalationInfo(complaint.id);
    this.displayViewModal = true;
  }

  closeViewModal() {
    this.displayViewModal = false;
    this.selectedComplaint = {};
    this.selectedAssigneeRole = '';
  }

  editComplaint(complaint: any) {
    if (!this.isSuperAdmin()) {
      alert('Only super admin can edit complaints.');
      return;
    }
    this.selectedComplaint = { ...complaint };
    this.originalAssignedTeam = complaint?.assignedTeam ?? '';
    this.loadEscalationInfo(complaint.id);
    if (this.isSuperAdmin()) {
      this.setAssignableRoles(complaint?.escalation?.source_app);
    }
    this.displayEditModal = true;
  }

  closeEditModal() {
    this.displayEditModal = false;
    this.selectedComplaint = {};
    this.originalAssignedTeam = '';
  }

  saveComplaint() {
    console.log('Saving complaint:', this.selectedComplaint);
    if (!this.canUpdateComplaint(this.selectedComplaint)) {
      alert('Only the currently assigned driver admin/supervisor can update this complaint.');
      return;
    }
    if (!this.selectedComplaint.id) {
      this.closeEditModal();
      return;
    }

    const requests: Observable<any>[] = [];

    const newAssignedTeam = this.normalizeRole(this.selectedComplaint.assignedTeam || '');
    const oldAssignedTeam = this.normalizeRole(this.originalAssignedTeam || this.escalationInfo?.current_team || '');

    if (this.isSuperAdmin() && newAssignedTeam && newAssignedTeam !== oldAssignedTeam) {
      const parsed = this.parseTeamName(newAssignedTeam);
      if (!parsed.role) {
        alert('Invalid assigned team selection.');
        return;
      }
      requests.push(this.complaintService.assignComplaint(this.selectedComplaint.id, parsed.role, parsed.scope));
    }

    if (this.selectedComplaint.activity) {
      requests.push(
        this.complaintService.updateComplaintStatus(
          this.selectedComplaint.id,
          'in_progress',
          this.selectedComplaint.activity
        )
      );
    }

    if (requests.length === 0) {
      this.closeEditModal();
      return;
    }

    forkJoin(requests).subscribe({
      next: () => {
        console.log('Complaint saved successfully');
        this.loadEscalationInfo(this.selectedComplaint.id);
        this.loadComplaints();
        this.closeEditModal();
      },
      error: (err) => {
        console.error('Error saving complaint:', err);
        alert('Failed to save complaint changes. Please try again.');
      }
    });
  }

  sendSolution() {
    console.log('Sending solution for:', this.selectedComplaint.id, this.solutionText);
    if (!this.canUpdateComplaint(this.selectedComplaint)) {
      alert('Only the currently assigned driver admin/supervisor can resolve this complaint.');
      return;
    }
    if (this.selectedComplaint.id && this.solutionText) {
      this.complaintService.updateComplaintStatus(
        this.selectedComplaint.id, 
        'resolved', 
        this.solutionText
      ).subscribe({
        next: () => {
          console.log('Solution sent successfully');
          this.loadComplaints();
        },
        error: (err) => console.error('Error sending solution:', err)
      });
    }
    this.closeViewModal();
  }

  loadComplaints(page: number = this.page) {
    this.isLoading = true;
    this.complaintService.loadComplaints(page, this.pageSize, this.getRoleFilterForTab()).subscribe({
      next: (response) => {
        this.page = response.page;
        this.totalRecords = response.total;
        this.totalPages = response.total_pages;
        this.complaints = response.data;
        this.rebuildAssignedTeamOptions();
        this.filterComplaints();
        this.updateStats();
        this.isLoading = false;
      },
      error: (err) => {
        console.error('Error loading complaints:', err);
        this.isLoading = false;
      }
    });
  }

  previousPage() {
    if (this.page > 1 && !this.isLoading) {
      this.loadComplaints(this.page - 1);
    }
  }

  nextPage() {
    if (this.page < this.totalPages && !this.isLoading) {
      this.loadComplaints(this.page + 1);
    }
  }

  onPageSizeChange() {
    this.page = 1;
    this.loadComplaints(1);
  }

  setActiveTab(tab: string) {
    this.activeTab = tab;
    this.page = 1;
    this.loadComplaints(1);
  }

  applyFilters() {
    let filtered = this.complaints;

    // Filter by role (tab)
    if (this.activeTab !== 'All') {
      const role = this.activeTab.endsWith('s') ? this.activeTab.slice(0, -1) : this.activeTab;
      filtered = filtered.filter(c => 
        c.role.includes(role) || 
        (role === 'Bus Owner' && c.role === 'Bus Owner') || 
        (role === 'Lounge owner' && c.role === 'Lounge owner')
      );
    }

    // Filter by category
    if (this.selectedCategory) {
      filtered = filtered.filter(c => c.category.toLowerCase() === this.selectedCategory.toLowerCase());
    }

    // Filter by status
    if (this.selectedStatus) {
      filtered = filtered.filter(c => c.status.toLowerCase() === this.selectedStatus.toLowerCase());
    }

    // Filter by assigned team
    if (this.selectedAssignedTeam) {
      filtered = filtered.filter(c => c.assignedTeam === this.selectedAssignedTeam);
    }

    // Filter by escalation level
    if (this.selectedLevel !== null) {
      filtered = filtered.filter(c => c.escalation?.current_level === this.selectedLevel);
    }

    // Filter by date
    if (this.selectedDate) {
      const filterDate = new Date(this.selectedDate);
      filterDate.setHours(0, 0, 0, 0);
      
      filtered = filtered.filter(c => {
        const complaintDate = new Date(c.dateTime);
        complaintDate.setHours(0, 0, 0, 0);
        return complaintDate.getTime() === filterDate.getTime();
      });
    }

    this.filteredComplaints = filtered;
  }

  clearFilters() {
    this.selectedCategory = '';
    this.selectedStatus = '';
    this.selectedAssignedTeam = '';
    this.selectedLevel = null;
    this.selectedDate = null;
    this.applyFilters();
  }

  filterComplaints() {
    this.applyFilters();
  }

  updateStats() {
    const total = this.totalRecords;
    const pending = this.complaints.filter(c => c.status.toLowerCase() === 'pending').length;
    const inProgress = this.complaints.filter(c => c.status.toLowerCase() === 'in progress').length;
    const resolved = this.complaints.filter(c => c.status.toLowerCase() === 'resolved').length;

    this.stats = [
      { title: 'Total Complaints', count: total, icon: 'pi pi-users', color: 'blue' },
      { title: 'Pending Complaints', count: pending, icon: 'pi pi-clock', color: 'indigo' },
      { title: 'Inprogress Complaints', count: inProgress, icon: 'pi pi-check-circle', color: 'purple' },
      { title: 'Resolved Complaints', count: resolved, icon: 'pi pi-check', color: 'green' }
    ];
  }

  getComplaintsByRole(role: string) {
    return this.complaints.filter(c => c.role === role);
  }

  getStatusSeverity(status: string): 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast' | undefined {
    switch (status.toLowerCase()) {
      case 'resolved': return 'success';
      case 'pending': return 'secondary';
      case 'in progress': return 'info';
      case 'escalated': return 'warn';
      default: return 'info';
    }
  }

  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  logout() {
    localStorage.removeItem('token');
    this.router.navigate(['/login']);
  }

  goUserProfile() {
    this.router.navigate(['/user-profile']);
  }

  loadEscalationInfo(complaintId: string) {
    this.complaintService.getComplaintEscalation(complaintId).subscribe({
      next: (data) => {
        this.escalationInfo = data;
        if (data?.current_team && this.selectedComplaint?.id) {
          this.selectedComplaint.assignedTeam = data.current_team;
        }
        if (this.isSuperAdmin() && data?.current_team) {
          this.selectedAssigneeRole = data.current_team;
        }
        if (this.isSuperAdmin()) {
          this.setAssignableRoles(data?.source_app);
        }
        console.log('Escalation info loaded:', data);
      },
      error: (err) => {
        console.error('Error loading escalation info:', err);
        this.escalationInfo = null;
      }
    });
  }

  escalateComplaint(complaintId: string) {
    if (!this.canEscalateComplaint) {
      alert('Only supervisors and super admins can escalate complaints.');
      return;
    }

    if (confirm('Are you sure you want to escalate this complaint to the next level?')) {
      this.isEscalating = true;
      this.complaintService.manualEscalateComplaint(complaintId, this.currentAdminId || 'manual').subscribe({
        next: (response) => {
          console.log('Complaint escalated successfully:', response);
          alert('Complaint escalated successfully! SMS notification sent to the next level team.');
          this.loadEscalationInfo(complaintId);
          this.loadComplaints();
          this.isEscalating = false;
        },
        error: (err) => {
          console.error('Error escalating complaint:', err);
          alert('Failed to escalate complaint. Please try again.');
          this.isEscalating = false;
        }
      });
    }
  }

  getEscalationBadgeClass(level: number): string {
    switch(level) {
      case 1: return 'level-1';
      case 2: return 'level-2';
      case 3: return 'level-3';
      default: return '';
    }
  }

  formatDate(dateString: string): string {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleString();
  }

  getResolutionWindowText(): string {
    const source = this.escalationInfo?.source_app;
    const level = this.escalationInfo?.current_level;
    if (source === 'driver' && (level === 1 || level === 2)) {
      return '5 days';
    }
    return level === 1 ? '48 hours' : '24 hours';
  }

  getTimeRemaining(nextDue?: string): string {
    if (!nextDue) {
      return 'N/A';
    }
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

  assignComplaintToSelectedRole() {
    if (!this.isSuperAdmin()) {
      alert('Only super admin can reassign complaints.');
      return;
    }
    if (!this.selectedComplaint?.id || !this.selectedAssigneeRole) {
      alert('Please select an app role.');
      return;
    }

    const parsed = this.parseTeamName(this.selectedAssigneeRole);
    if (!parsed.role) {
      alert('Invalid role selection.');
      return;
    }

    this.isAssigningComplaint = true;
    this.complaintService.assignComplaint(this.selectedComplaint.id, parsed.role, parsed.scope).subscribe({
      next: () => {
        alert('Complaint assigned by app role successfully.');
        this.loadEscalationInfo(this.selectedComplaint.id);
        this.loadComplaints();
        this.isAssigningComplaint = false;
      },
      error: (err) => {
        console.error('Error assigning complaint:', err);
        alert('Failed to assign complaint.');
        this.isAssigningComplaint = false;
      }
    });
  }

  private setAssignableRoles(sourceApp?: string) {
    const normalizedScope = this.normalizeRole(sourceApp || this.escalationInfo?.source_app || '');
    const scopes = ['bus', 'driver', 'lounges', 'passenger'];

    const toLabel = (value: string) => value.replace(/_/g, ' ').replace(/\b\w/g, ch => ch.toUpperCase());

    const values = normalizedScope && scopes.includes(normalizedScope)
      ? [`${normalizedScope}_admin`, `${normalizedScope}_supervisor`, 'super_admin']
      : [
          'bus_admin', 'bus_supervisor',
          'driver_admin', 'driver_supervisor',
          'lounges_admin', 'lounges_supervisor',
          'passenger_admin', 'passenger_supervisor',
          'super_admin'
        ];

    this.assignableRoles = values.map(value => ({ label: toLabel(value), value }));
  }

  private rebuildAssignedTeamOptions() {
    const teams = Array.from(new Set(this.complaints.map(c => c.assignedTeam).filter(Boolean))).sort();
    this.assignedTeamOptions = [
      { label: 'All Teams', value: '' },
      ...teams.map(team => ({ label: team, value: team }))
    ];
  }

  private normalizeRole(role: string): string {
    return (role || '').toLowerCase().trim().replace(/[-\s]+/g, '_');
  }

  private getRoleFilterForTab(): string {
    switch (this.activeTab) {
      case 'Passengers':
        return 'passenger';
      case 'Drivers':
        return 'driver';
      case 'Conductors':
        return 'conductor';
      case 'Bus Owners':
        return 'bus_owner';
      case 'Lounge owners':
        return 'lounge_owner';
      default:
        return '';
    }
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

    if (scope === 'driver') {
      return role === 'admin' ? 'driver_supervisor' : 'driver_admin';
    }

    return `${scope}_${role}`;
  }

  private parseTeamName(team: string): { role: string; scope: string } {
    const normalized = this.normalizeRole(team);
    if (normalized === 'super_admin') {
      return { role: 'super_admin', scope: '' };
    }

    const idx = normalized.lastIndexOf('_');
    if (idx <= 0 || idx === normalized.length - 1) {
      return { role: '', scope: '' };
    }

    return {
      scope: normalized.substring(0, idx),
      role: normalized.substring(idx + 1)
    };
  }

  isSuperAdmin(): boolean {
    return this.normalizeRole(this.currentAdminRole) === 'super_admin';
  }
}
