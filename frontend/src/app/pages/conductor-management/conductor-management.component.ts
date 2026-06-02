import { Component, OnInit, AfterViewInit, Inject, PLATFORM_ID, ViewChild, ElementRef } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';

import Chart from 'chart.js/auto';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { ConductorService } from '../../core/services/conductor.service';
import { Conductor } from '../../core/models/conductor.model';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-conductor-management',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent, NotificationPanelComponent, RouterModule],
  templateUrl: './conductor-management.component.html',
  styleUrls: ['./conductor-management.component.scss']
})
export class ConductorManagementComponent implements OnInit, AfterViewInit {
  conductors: Conductor[] = [];
  filteredConductors: Conductor[] = [];
  searchTerm: string = '';
  statusFilter: string = 'All';
  experienceLevels = ['0-2yrs', '3-5yrs', '6-10yrs', '10+yrs'];
  showNotificationPanel = false;
  showProfileMenu = false;

  showAddConductorModal = false;
  showEditConductorModal = false;

  newConductor: Omit<Conductor, 'id'> = {
    name: '',
    contact_number: '',
    experience_years: 0,
    license_number: '',
    license_expiry_date: new Date().toISOString().split('T')[0],
    verification_status: 'pending',
    verification_notes: '',
    status: 'active',
    hire_date: new Date().toISOString().split('T')[0]
  };

  selectedConductor: Conductor | null = null;

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // Chart properties
  barChartData: any;
  barChartOptions: any;
  isBrowser: boolean;
  @ViewChild('barChartCanvas', { static: false }) barChartCanvas!: ElementRef<HTMLCanvasElement>;
  chart: Chart | null = null;

  constructor(private router: Router, private conductorService: ConductorService, @Inject(PLATFORM_ID) private platformId: Object, public notificationService: NotificationService) {
    this.isBrowser = isPlatformBrowser(this.platformId);
  }

  ngOnInit(): void {
    this.conductorService.conductors$.subscribe(conductors => {
      this.conductors = conductors;
      this.filteredConductors = conductors;
      if (this.isBrowser) {
        this.updateBarChart();
      }
      // Apply current sorting to initial data
      if (this.sortColumn) {
        this.applySorting();
      }
    });
  }

  addConductor() {
    this.showAddConductorModal = true;
  }

  closeAddConductorModal() {
    this.showAddConductorModal = false;
    this.newConductor = {
      name: '',
      contact_number: '',
      experience_years: 0,
      license_number: '',
      license_expiry_date: new Date().toISOString().split('T')[0],
      verification_status: 'pending',
      verification_notes: '',
      status: 'active',
      hire_date: new Date().toISOString().split('T')[0]
    };
  }

  // Validate conductor form - all required fields must be filled
  isConductorFormValid(): boolean {
    return !!(
      this.newConductor.name?.trim() &&
      this.newConductor.contact_number?.trim() &&
      this.newConductor.license_number?.trim() &&
      this.newConductor.experience_years >= 0 &&
      this.newConductor.license_expiry_date &&
      this.newConductor.hire_date
    );
  }

  saveConductor() {
    // Validate all required fields
    if (!this.isConductorFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }

    if (this.newConductor.name && this.newConductor.contact_number) {
      this.conductorService.addConductor({
        name: this.newConductor.name,
        contact_number: this.newConductor.contact_number,
        experience_years: this.newConductor.experience_years,
        license_number: this.newConductor.license_number,
        license_expiry_date: this.newConductor.license_expiry_date,
        verification_status: this.newConductor.verification_status,
        verification_notes: this.newConductor.verification_notes,
        status: this.newConductor.status,
        hire_date: this.newConductor.hire_date,
        id: '' // Backend will generate
      }).subscribe({
        next: () => {
          console.log('✓ Conductor added successfully');
          alert('Conductor added successfully');
          this.conductorService.loadConductors(); // Reload conductors to refresh the list
          this.closeAddConductorModal();
        },
        error: (err) => {
          console.error('✗ Failed to add conductor:', err);
          alert(`Failed to add conductor: ${err.error?.error || err.message || 'Unknown error'}`);
        }
      });
    } else {
      alert('Please fill all required fields');
    }
  }

  updateConductor(conductor: Conductor) {
    this.selectedConductor = { ...conductor };
    this.showEditConductorModal = true;
  }

  closeEditConductorModal() {
    this.showEditConductorModal = false;
    this.selectedConductor = null;
  }

  saveEditConductor() {
    if (this.selectedConductor) {
      this.conductorService.updateConductor(this.selectedConductor).subscribe({
        next: () => {
          this.conductorService.loadConductors(); // Reload conductors to refresh the list
          this.closeEditConductorModal();
        },
        error: (err) => console.error('Failed to update conductor', err)
      });
    }
  }

  toggleStatus(conductor: Conductor) {
    const statusOptions: Array<string> = ['Active', 'Inactive', 'Resigned'];
    const currentIndex = statusOptions.indexOf(conductor.status);
    const newStatus = statusOptions[(currentIndex + 1) % statusOptions.length];
    
    // Update the conductor object with new status
    const updatedConductor = { ...conductor, status: newStatus };
    
    // Save to backend
    this.conductorService.updateConductor(updatedConductor).subscribe({
      next: (updated) => {
        console.log(`${conductor.name} status updated to ${newStatus}`);
        this.conductorService.loadConductors(); // Reload to reflect changes
      },
      error: (err) => {
        console.error('Failed to update conductor:', err);
        alert(`Failed to update conductor: ${err.error?.error || err.message}`);
        this.conductorService.loadConductors(); // Reload to revert UI changes
      }
    });
  }

  deleteConductor(conductor: Conductor) {
    const confirmed = confirm(`Are you sure you want to delete ${conductor.name}?`);
    if (confirmed) {
      this.conductorService.deleteConductor(conductor.id).subscribe({
        next: () => {
          console.log(`${conductor.name} deleted`);
          this.conductorService.loadConductors(); // Reload conductors to refresh the list
        },
        error: (err) => console.error('Failed to delete conductor', err)
      });
    }
  }

  exportConductorsPdf() {
    const doc = new jsPDF({ orientation: 'landscape' });
    doc.setFontSize(16);
    doc.text('Conductors Management Report', 14, 16);

    const tableHead = [['Conductor ID', 'Conductor Name', 'Contact', 'License Number', 'License Expire Date', 'Experience Yrs', 'Verification', 'Verification Note', 'Status']];
    const tableBody = this.filteredConductors.map(c => [
      c.id,
      c.name,
      c.contact_number,
      `${c.experience_years}yrs`,
      c.license_number,
      new Date(c.license_expiry_date).toLocaleDateString(),
      c.verification_status,
      c.verification_notes || '-',
      c.status,
    ]);

    autoTable(doc, {
      head: tableHead,
      body: tableBody,
      startY: 22,
      styles: { fontSize: 10 },
      headStyles: { fillColor: [59, 130, 246] }
    });

    doc.save('conductors-report.pdf');
  }

  // Stats helpers
  getTotalConductors(): number { return this.conductors.length; }
  getActiveCount(): number { return this.conductors.filter(c => c.status.toLowerCase() === 'active').length; }
  getInactiveCount(): number { return this.conductors.filter(c => c.status.toLowerCase() === 'inactive').length; }
  getResignedCount(): number { return this.conductors.filter(c => c.status.toLowerCase() === 'resigned').length; }
  getAverageExperience(): number {
    if (this.conductors.length === 0) return 0;
    const total = this.conductors.reduce((sum, c) => sum + c.experience_years, 0);
    return Math.round(total / this.conductors.length);
  }
  //notifications 
    toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }
  
  // Filtered stats helpers for pie chart
  getFilteredTotalConductors(): number { return this.filteredConductors.length; }
  getFilteredActiveCount(): number { return this.filteredConductors.filter(c => c.status.toLowerCase() === 'active').length; }
  // Treat any non-active status as Inactive for charts
  getFilteredInactiveCount(): number { return this.filteredConductors.filter(c => c.status.toLowerCase() !== 'active').length; }

  // Search and filter functionality
  onSearchChange(): void {
    this.applyFilters();
  }

  onStatusFilterChange(): void {
    this.applyFilters();
  }

  private applyFilters(): void {
    let filtered = this.conductors;

    // Apply status filter
    if (this.statusFilter !== 'All') {
      filtered = filtered.filter(c => c.status.toLowerCase() === this.statusFilter.toLowerCase());
    }

    // Apply search filter - search across all columns
    if (this.searchTerm.trim()) {
      const searchLower = this.searchTerm.toLowerCase();
      filtered = filtered.filter(conductor =>
        conductor.name.toLowerCase().includes(searchLower) ||
        conductor.contact_number.toLowerCase().includes(searchLower) ||
        conductor.id.toLowerCase().includes(searchLower) ||
        conductor.experience_years.toString().includes(this.searchTerm) ||
        conductor.license_number.toLowerCase().includes(searchLower) ||
        conductor.license_expiry_date.toLowerCase().includes(searchLower) ||
        (conductor.verification_status?.toLowerCase() || '').includes(searchLower) ||
        (conductor.verification_notes?.toLowerCase() || '').includes(searchLower) ||
        conductor.status.toLowerCase().includes(searchLower) ||
        conductor.hire_date.toLowerCase().includes(searchLower)
      );
    }

    this.filteredConductors = filtered;
  }

  clearSearch(): void {
    this.searchTerm = '';
    this.applyFilters();
  }

  clearFilters(): void {
    this.searchTerm = '';
    this.statusFilter = 'All';
    this.filteredConductors = this.conductors;
  }

  // Navigation methods
  goDashboard(): void { this.router.navigate(['/dashboard']); }
  goBusManagement(): void { this.router.navigate(['/bus-management']); }
  goPassengerManagement(): void { this.router.navigate(['/passenger-management']); }
  goLounges(): void { this.router.navigate(['/lounges']); }

  // Experience level calculations for charts
  getExperienceLevelCount(level: string): number {
    switch (level) {
      case '0-2yrs': return this.conductors.filter(c => c.experience_years >= 0 && c.experience_years <= 2).length;
      case '3-5yrs': return this.conductors.filter(c => c.experience_years >= 3 && c.experience_years <= 5).length;
      case '6-10yrs': return this.conductors.filter(c => c.experience_years >= 6 && c.experience_years <= 10).length;
      case '10+yrs': return this.conductors.filter(c => c.experience_years > 10).length;
      default: return 0;
    }
  }

  getExperienceLevelPercentage(level: string): number {
    const total = this.conductors.length || 1;
    return (this.getExperienceLevelCount(level) / total) * 100;
  }

  // Sorting functionality
  onSort(column: string): void {
    if (this.sortColumn === column) {
      // Toggle direction if same column
      this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      // New column, start with ascending
      this.sortColumn = column;
      this.sortDirection = 'asc';
    }
    this.applySorting();
  }

  private applySorting(): void {
    this.filteredConductors = [...this.filteredConductors].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (this.sortColumn) {
        case 'id':
          aValue = a.id.toLowerCase();
          bValue = b.id.toLowerCase();
          break;
        case 'name':
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
        case 'experience_years':
          aValue = a.experience_years;
          bValue = b.experience_years;
          break;
        case 'license_expiry_date':
          aValue = new Date(a.license_expiry_date).getTime();
          bValue = new Date(b.license_expiry_date).getTime();
          break;
        case 'hire_date':
          aValue = new Date(a.hire_date).getTime();
          bValue = new Date(b.hire_date).getTime();
          break;
        case 'verification_status':
          aValue = (a.verification_status || '').toLowerCase();
          bValue = (b.verification_status || '').toLowerCase();
          break;
        case 'status':
          aValue = a.status.toLowerCase();
          bValue = b.status.toLowerCase();
          break;
        default:
          return 0;
      }

      if (aValue < bValue) {
        return this.sortDirection === 'asc' ? -1 : 1;
      }
      if (aValue > bValue) {
        return this.sortDirection === 'asc' ? 1 : -1;
      }
      return 0;
    });
  }

  getSortIcon(column: string): string {
    if (this.sortColumn !== column) {
      return ' ⇅'; // Both arrows for unsorted columns
    }
    return this.sortDirection === 'asc' ? ' ↑' : ' ↓';
  }

  getPieChartGradient(): string {
    const total = this.getFilteredTotalConductors();
    if (total === 0) return 'conic-gradient(gray 0deg 360deg)';

    const activeCount = this.getFilteredActiveCount();
    const inactiveCount = this.getFilteredInactiveCount();

    const activePercent = (activeCount / total) * 360;
    const inactivePercent = (inactiveCount / total) * 360;

    const activeEnd = activePercent;
    const inactiveEnd = activeEnd + inactivePercent;

    return `conic-gradient(var(--active) 0deg ${activeEnd}deg, var(--inactive) ${activeEnd}deg ${inactiveEnd}deg)`;
  }

  ngAfterViewInit(): void {
    if (this.isBrowser) {
      this.createChart();
    }
  }

  createChart(): void {
    if (this.barChartCanvas && this.barChartCanvas.nativeElement) {
      const ctx = this.barChartCanvas.nativeElement.getContext('2d');
      if (ctx) {
        this.chart = new Chart(ctx, {
          type: 'bar',
          data: {
            labels: ['0-2yrs', '3-5yrs', '6-10yrs', '10+yrs'],
            datasets: [{
              data: [
                this.getExperienceLevelCount('0-2yrs'),
                this.getExperienceLevelCount('3-5yrs'),
                this.getExperienceLevelCount('6-10yrs'),
                this.getExperienceLevelCount('10+yrs')
              ],
              backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
              borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
              borderWidth: 0.25
            }]
          },
          options: {
            responsive: true,
            plugins: {
              legend: {
                display: false
              }
            },
            scales: {
              y: {
                beginAtZero: true,
                ticks: {
                  stepSize: 1
                }
              }
            }
          }
        });
      }
    }
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('admin_user');
    this.router.navigate(['/login']);
  }

  updateBarChart(): void {
    if (this.chart) {
      this.chart.data.datasets[0].data = [
        this.getExperienceLevelCount('0-2yrs'),
        this.getExperienceLevelCount('3-5yrs'),
        this.getExperienceLevelCount('6-10yrs'),
        this.getExperienceLevelCount('10+yrs')
      ];
      this.chart.update();
    }
  }
}
