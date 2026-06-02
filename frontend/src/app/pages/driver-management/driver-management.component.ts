import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { BaseChartDirective } from 'ng2-charts';
import { Chart, registerables } from 'chart.js';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { DriverService } from '../../core/services/driver.service';
import { Driver } from '../../core/models/driver.model';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-driver-management',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent, BaseChartDirective, NotificationPanelComponent, RouterModule],
  templateUrl: './driver-management.component.html',
  styleUrls: ['./driver-management.component.scss']
})
export class DriverManagementComponent implements OnInit {
  drivers: Driver[] = [];
  filteredDrivers: Driver[] = [];
  searchTerm: string = '';
  statusFilter: 'All' | 'Active' | 'Inactive' = 'All';
  experienceFilter: string = 'All';
  experienceLevels = ['0-2yrs', '3-5yrs', '6-10yrs', '10+yrs'];
  showNotificationPanel = false;
  showProfileMenu = false;

  showAddDriverModal = false;
  isEditing = false;
  editingDriver: Driver | null = null;
  driverName: string = '';

  newDriver: Omit<Driver, 'id'> = {
    name: '',
    contact_number: '',
    license_number: '',
    experience_years: 0,
    status: 'Active',
    license_expiry_date: '',
    verification_status: 'pending',
    verification_notes: '',
    hire_date: ''
  };

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // Chart properties
  barChartData: any;
  barChartOptions: any;
  isBrowser: boolean;

  constructor(private router: Router, private driverService: DriverService, @Inject(PLATFORM_ID) private platformId: Object, public notificationService: NotificationService) {
    this.isBrowser = isPlatformBrowser(this.platformId);
  }

  ngOnInit(): void {
    if (this.isBrowser) {
      Chart.register(...registerables);
    }
    this.driverService.drivers$.subscribe(drivers => {
      this.drivers = drivers;
      this.filteredDrivers = drivers;
      if (this.isBrowser) {
        this.updateBarChart();
      }
      // Apply current sorting to initial data
      if (this.sortColumn) {
        this.applySorting();
      }
    });
  }

  addDriver() {
    this.isEditing = false;
    this.editingDriver = null;
    this.driverName = '';
    this.showAddDriverModal = true;
  }

  closeAddDriverModal() {
    this.showAddDriverModal = false;
    this.isEditing = false;
    this.editingDriver = null;
    this.driverName = '';
    this.newDriver = {
      name: '',
      contact_number: '',
      license_number: '',
      experience_years: 0,
      status: 'Active',
      license_expiry_date: '',
      verification_status: 'pending',
      verification_notes: '',
      hire_date: ''
    };
  }

  // Validate driver form - all required fields must be filled
  isDriverFormValid(): boolean {
    return !!(
      this.newDriver.name?.trim() &&
      this.newDriver.contact_number?.trim() &&
      this.newDriver.license_number?.trim() &&
      this.newDriver.experience_years >= 0 &&
      this.newDriver.license_expiry_date &&
      this.newDriver.hire_date
    );
  }

  saveDriver() {
    // Trim all text fields
    this.newDriver.name = this.newDriver.name?.trim() || '';
    this.newDriver.contact_number = this.newDriver.contact_number?.trim() || '';
    this.newDriver.license_number = this.newDriver.license_number?.trim() || '';

    // Validate required fields
    const requiredFields = [
      { name: 'Driver Name', value: this.newDriver.name },
      { name: 'Contact Number', value: this.newDriver.contact_number },
      { name: 'License Number', value: this.newDriver.license_number }
    ];
    
    const missingFields = requiredFields.filter(f => !f.value);
    if (missingFields.length > 0) {
      console.log('Missing fields:', missingFields);
      console.log('Current driver data:', this.newDriver);
      alert(`Please fill all required fields: ${missingFields.map(f => f.name).join(', ')}`);
      return;
    }

    if (this.newDriver.experience_years < 0) {
      alert('Experience years must be 0 or greater');
      return;
    }

    if (this.isEditing && this.editingDriver) {
      const updatedDriver = { ...this.editingDriver, ...this.newDriver };
      console.log('Updating driver:', updatedDriver);
      this.driverService.updateDriver(updatedDriver).subscribe({
        next: (response) => {
          console.log('✓ Driver updated successfully:', response);
          alert('Driver updated successfully');
          this.driverService.loadDrivers(); // Reload drivers to refresh the list
          this.closeAddDriverModal();
        },
        error: (err) => {
          console.error('✗ Failed to update driver:', err);
          alert(`Failed to update driver: ${err.error?.error || err.message || 'Unknown error'}`);
        }
      });
    } else {
      const driverToAdd: Driver = { ...this.newDriver, id: '' };
      console.log('Adding new driver:', driverToAdd);
      this.driverService.addDriver(driverToAdd).subscribe({
        next: (response) => {
          console.log('✓ Driver added successfully:', response);
          alert('Driver added successfully');
          this.driverService.loadDrivers(); // Reload drivers to refresh the list
          this.closeAddDriverModal();
        },
        error: (err) => {
          console.error('✗ Failed to add driver:', err);
          alert(`Failed to add driver: ${err.error?.error || err.message || 'Unknown error'}`);
        }
      });
    }
  }

  goDashboard() {
    this.router.navigate(['/dashboard']);
  }

  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

  updateDriver(driver: Driver) {
    this.isEditing = true;
    this.editingDriver = driver;
    this.driverName = driver.name;
    this.newDriver = {
      name: driver.name,
      contact_number: driver.contact_number,
      license_number: driver.license_number,
      experience_years: driver.experience_years,
      status: driver.status,
      license_expiry_date: driver.license_expiry_date,
      hire_date: driver.hire_date,
      verification_status: driver.verification_status,
      verification_notes: driver.verification_notes
    };
    this.showAddDriverModal = true;
  }

  toggleActive(driver: Driver) {
    const currentStatus = driver.status.toLowerCase();
    const newStatus = currentStatus === 'active' ? 'Inactive' : 'Active';
    
    // Update the driver object with new status
    const updatedDriver = { ...driver, status: newStatus };
    
    // Save to backend
    this.driverService.updateDriver(updatedDriver).subscribe({
      next: (updated) => {
        console.log(`${driver.name} status updated to ${newStatus}`);
        this.driverService.loadDrivers(); // Reload to reflect changes
      },
      error: (err) => {
        console.error('Failed to update driver:', err);
        alert(`Failed to update driver: ${err.error?.error || err.message}`);
        this.driverService.loadDrivers(); // Reload to revert UI changes
      }
    });
  }

  deleteDriver(driver: Driver) {
    const confirmed = confirm(`Are you sure you want to delete ${driver.name}?`);
    if (confirmed) {
      this.driverService.deleteDriver(driver.id).subscribe({
        next: () => {
          console.log(`${driver.name} deleted`);
          this.driverService.loadDrivers(); // Reload drivers to refresh the list
        },
        error: (err) => console.error('Failed to delete driver', err)
      });
    }
  }

  exportDriversPdf() {
    const doc = new jsPDF({ orientation: 'landscape' });
    doc.setFontSize(16);
    doc.text('Drivers History', 14, 16);

    const tableHead = [['Driver ID', 'Name', 'Contact', 'License No.', 'Experience', 'Status']];
    const tableBody = this.drivers.map(d => [
      d.id,
      d.name,
      this.formatPhone(d.contact_number),
      d.license_number,
      `${d.experience_years}yrs`,
      d.status
    ]);

    autoTable(doc, {
      head: tableHead,
      body: tableBody,
      startY: 22,
      styles: { fontSize: 10 },
      headStyles: { fillColor: [59, 130, 246] }
    });

    doc.save('drivers-history.pdf');
  }

  exportCSV() {
    const headers = ['Driver ID', 'Name', 'Contact', 'License No.', 'Experience', 'Status', 'License Expiry', 'Hire Date', 'Verification Status'];
    const rows = this.filteredDrivers.map(d => [
      d.id,
      d.name,
      this.formatPhone(d.contact_number),
      d.license_number,
      `${d.experience_years}yrs`,
      d.status,
      d.license_expiry_date,
      d.hire_date,
      d.verification_status
    ].join(','));

    const csvContent = [headers.join(','), ...rows].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', 'drivers.csv');
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  // Stats helpers
  getTotalDrivers(): number {
    return this.drivers.length;
  }

  getActiveCount(): number {
    return this.drivers.filter(driver => driver.status.toLowerCase() === 'active').length;
  }

  getInactiveCount(): number {
    return this.drivers.filter(driver => driver.status.toLowerCase() !== 'active').length;
  }

  getAverageExperience(): number {
    if (this.drivers.length === 0) return 0;
    const total = this.drivers.reduce((sum, driver) => sum + driver.experience_years, 0);
    return Math.round(total / this.drivers.length);
  }

  getActivePercentage(): number {
    const total = this.drivers.length || 1;
    return (this.getActiveCount() / total) * 100;
  }

  getInactivePercentage(): number {
    const total = this.drivers.length || 1;
    return (this.getInactiveCount() / total) * 100;
  }

  getDriverPieBackground(): string {
    const active = this.getActivePercentage();
    return `conic-gradient(#0046FF 0% ${active}%, #FAA533 ${active}% 100%)`;
  }

  getExperienceLevelCount(level: string): number {
    switch (level) {
      case '0-2yrs': return this.drivers.filter(d => d.experience_years >= 0 && d.experience_years <= 2).length;
      case '3-5yrs': return this.drivers.filter(d => d.experience_years >= 3 && d.experience_years <= 5).length;
      case '6-10yrs': return this.drivers.filter(d => d.experience_years >= 6 && d.experience_years <= 10).length;
      case '10+yrs': return this.drivers.filter(d => d.experience_years > 10).length;
      default: return 0;
    }
  }

  getExperienceLevelPercentage(level: string): number {
    const total = this.drivers.length || 1;
    return (this.getExperienceLevelCount(level) / total) * 100;
  }

  // Search and filter functionality
  filterDrivers() {
    let tempDrivers = this.drivers;

    // Filter by status
    if (this.statusFilter !== 'All') {
      tempDrivers = tempDrivers.filter(driver => driver.status === this.statusFilter);
    }

    // Filter by experience
    if (this.experienceFilter !== 'All') {
      const [min, max] = this.experienceFilter.replace('yrs', '').replace('+', '-Infinity').split('-').map(Number);
      tempDrivers = tempDrivers.filter(driver => {
        const exp = driver.experience_years;
        if (max === Infinity) {
          return exp >= min;
        }
        return exp >= min && exp <= max;
      });
    }

    // Filter by search term
    if (this.searchTerm) {
      const lowercasedTerm = this.searchTerm.toLowerCase();
      tempDrivers = tempDrivers.filter(driver =>
        driver.name.toLowerCase().includes(lowercasedTerm) ||
        driver.contact_number.includes(lowercasedTerm) ||
        driver.license_number.toLowerCase().includes(lowercasedTerm)
      );
    }

    this.filteredDrivers = tempDrivers;
    this.applySorting();
  }

  onSearchChange(): void {
    this.filterDrivers();
  }

  onFilterChange(): void {
    this.filterDrivers();
  }

  clearFilters(): void {
    this.searchTerm = '';
    this.statusFilter = 'All';
    this.experienceFilter = 'All';
    this.filterDrivers();
  }

  clearSearch(): void {
    this.searchTerm = '';
    this.filterDrivers();
  }

  goBusManagement(): void {
    this.router.navigate(['/bus-management']);
  }

  goPassengerManagement(): void {
    this.router.navigate(['/passenger-management']);
  }

  goLounges(): void {
    this.router.navigate(['/lounges']);
  }

  goPassengerReports(): void {
    this.router.navigate(['/scheduling']);
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
    this.filteredDrivers = [...this.filteredDrivers].sort((a, b) => {
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
        case 'license_expiry_date':
          aValue = new Date(a.license_expiry_date || '1970-01-01').getTime();
          bValue = new Date(b.license_expiry_date || '1970-01-01').getTime();
          break;
        case 'experience_years':
          aValue = a.experience_years;
          bValue = b.experience_years;
          break;
        case 'hire_date':
          aValue = new Date(a.hire_date || '1970-01-01').getTime();
          bValue = new Date(b.hire_date || '1970-01-01').getTime();
          break;
        case 'verification_status':
          aValue = a.verification_status || '';
          bValue = b.verification_status || '';
          break;
        case 'status':
          aValue = a.status;
          bValue = b.status;
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

  formatPhone(phone: string): string {
    return `(${phone})`;
  }

  private generateDriverId(): string {
    const existingIds = this.drivers.map(d => d.id);
    let counter = 1;
    let newId = `DRV${counter.toString().padStart(3, '0')}`;
    while (existingIds.includes(newId)) {
      counter++;
      newId = `DRV${counter.toString().padStart(3, '0')}`;
    }
    return newId;
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
    this.barChartData = {
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
    };
    this.barChartOptions = {
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
    };
  }
}


