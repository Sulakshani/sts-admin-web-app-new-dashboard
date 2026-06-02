import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule, ActivatedRoute } from '@angular/router';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { BaseChartDirective } from 'ng2-charts';
import { Chart, registerables } from 'chart.js';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { BusService } from '../../core/services/bus.service';
import { NotificationService } from '../../core/services/notification.service';


import { Bus } from '../../core/models/bus.model';

@Component({
  selector: 'app-bus-management',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent, BaseChartDirective, NotificationPanelComponent, RouterModule],
  templateUrl: './bus-management.component.html',
  styleUrls: ['./bus-management.component.scss']
})
export class BusManagementComponent implements OnInit {
  buses: Bus[] = [];
  filteredBuses: Bus[] = [];
  searchTerm: string = '';
  searchCriteria: string = 'bus_number'; // Default search criteria
  statusFilter: 'all' | 'active' | 'inactive' = 'all';
  currentPage: string = 'bus-management';
  showNotificationPanel = false;
  showProfileMenu = false;

  showAddBusModal = false;
  showEditBusModal = false;

  newBus: Omit<Bus, 'id'> = {
    bus_number: '',
    company_name: '',
    identify_or_incorporation_no: '',
    business_email: '',
    business_phone: '',
    permit_number: '',
    license_plate: '',
    total_seats: 0,
    bus_type: 'Standard',
    custom_route_name: '',
    fare_per_seat: 0,
    status: 'active',
    verification_status: 'Pending',
    verification_documents: []
  };

  selectedBus: Bus | null = null;

  // Document upload properties
  showDocumentModal = false;
  selectedBusForDocuments: Bus | null = null;
  selectedFiles: File[] = [];
  selectedDocumentType: string = '';

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // Chart properties
  barChartData: any;
  barChartOptions: any;
  isBrowser: boolean;
  private pendingEditBusId: string | null = null;

  constructor(private router: Router, private route: ActivatedRoute, private busService: BusService, @Inject(PLATFORM_ID) private platformId: Object, public notificationService: NotificationService) {
    this.isBrowser = isPlatformBrowser(this.platformId);
  }

  ngOnInit(): void {
    if (this.isBrowser) {
      Chart.register(...registerables);
    }
    this.busService.buses$.subscribe(buses => {
      console.log('Buses loaded in component:', buses);
      this.buses = buses;
      this.applyFilters();
      this.tryOpenEditBusModalFromQuery();
      if (this.isBrowser) {
        this.updateBarChart();
      }
    });

    this.route.queryParams.subscribe(params => {
      this.pendingEditBusId = params['editBusId'] || null;
      this.tryOpenEditBusModalFromQuery();
    });
  }

  private tryOpenEditBusModalFromQuery(): void {
    if (!this.pendingEditBusId || this.showEditBusModal) {
      return;
    }

    const busToEdit = this.buses.find(bus => bus.id === this.pendingEditBusId);
    if (!busToEdit) {
      return;
    }

    this.selectedBus = { ...busToEdit };
    this.showEditBusModal = true;
    this.pendingEditBusId = null;

    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: { editBusId: null },
      queryParamsHandling: 'merge',
      replaceUrl: true
    });
  }

 

  goDashboard() {
    this.router.navigate(['/dashboard']);
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

  addBus() {
    this.showAddBusModal = true;
  }

  closeAddBusModal() {
    this.showAddBusModal = false;
    this.newBus = {
      bus_number: '',
      company_name: '',
      identify_or_incorporation_no: '',
      business_email: '',
      business_phone: '',
      permit_number: '',
      license_plate: '',
      total_seats: 0,
      bus_type: 'Standard',
      custom_route_name: '',
      fare_per_seat: 0,
      status: 'active',
      verification_status: 'Pending',
      verification_documents: []
    };
  }

  // Validate bus form - all required fields must be filled
  isBusFormValid(): boolean {
    return !!(
      this.newBus.bus_number?.trim() &&
      this.newBus.company_name?.trim() &&
      this.newBus.identify_or_incorporation_no?.trim() &&
      this.newBus.business_email?.trim() &&
      this.newBus.business_phone?.trim() &&
      this.newBus.permit_number?.trim() &&
      this.newBus.license_plate?.trim() &&
      this.newBus.custom_route_name?.trim() &&
      this.newBus.total_seats > 0 &&
      this.newBus.fare_per_seat >= 0
    );
  }

  saveBus() {
    // Validate all required fields
    if (!this.isBusFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }
    
    // Ensure verification_documents is an array
    if (!this.newBus.verification_documents) {
      this.newBus.verification_documents = [];
    }
    
    // Create a clean copy of the bus object
    const busData = {
      bus_number: this.newBus.bus_number,
      company_name: this.newBus.company_name,
      identify_or_incorporation_no: this.newBus.identify_or_incorporation_no,
      business_email: this.newBus.business_email,
      business_phone: this.newBus.business_phone,
      permit_number: this.newBus.permit_number,
      license_plate: this.newBus.license_plate,
      total_seats: this.newBus.total_seats,
      bus_type: this.newBus.bus_type,
      custom_route_name: this.newBus.custom_route_name,
      fare_per_seat: this.newBus.fare_per_seat,
      status: this.newBus.status,
      verification_status: this.newBus.verification_status,
      verification_documents: this.newBus.verification_documents || []
    };
    
    console.log('Sending bus data to API:', busData);
    this.busService.addBus(busData).subscribe({
      next: (response) => {
        console.log('✓ Bus added successfully:', response);
        // No need to call loadBuses() - the service already updates the local state via tap operator
        this.closeAddBusModal();
        alert('Bus added successfully!');
      },
      error: (err) => {
        console.error('✗ Failed to add bus:', err);
        
        // Parse the error message for user-friendly display
        let errorMessage = 'Unknown error';
        if (err.error?.error) {
          const errorText = err.error.error;
          if (errorText.includes('buses_license_plate_key')) {
            errorMessage = `License plate "${busData.license_plate}" already exists. Please use a different registration number.`;
          } else if (errorText.includes('buses_bus_number_key')) {
            errorMessage = `Bus number "${busData.bus_number}" already exists. Please use a different bus number.`;
          } else if (errorText.includes('buses_permit_number_key')) {
            errorMessage = `Permit number "${busData.permit_number}" already exists. Please use a different permit number.`;
          } else if (errorText.includes('duplicate key')) {
            errorMessage = 'This bus details already exist in the system. Please check your input.';
          } else {
            errorMessage = errorText;
          }
        } else if (err.message) {
          errorMessage = err.message;
        }
        
        alert(`Failed to add bus: ${errorMessage}`);
      }
    });
  }

  closeEditBusModal() {
    this.showEditBusModal = false;
    this.selectedBus = null;
  }

  saveEditBus() {
    if (!this.selectedBus) {
      alert('No bus selected');
      return;
    }

    // Validate required fields
    const requiredFields = [
      { name: 'bus_number', value: this.selectedBus.bus_number },
      { name: 'company_name', value: this.selectedBus.company_name },
      { name: 'identify_or_incorporation_no', value: this.selectedBus.identify_or_incorporation_no },
      { name: 'business_email', value: this.selectedBus.business_email },
      { name: 'business_phone', value: this.selectedBus.business_phone },
      { name: 'permit_number', value: this.selectedBus.permit_number },
      { name: 'license_plate', value: this.selectedBus.license_plate },
      { name: 'custom_route_name', value: this.selectedBus.custom_route_name }
    ];
    
    const missingFields = requiredFields.filter(f => !f.value);
    if (missingFields.length > 0) {
      alert(`Please fill all required fields: ${missingFields.map(f => f.name).join(', ')}`);
      return;
    }
    
    if (this.selectedBus.total_seats <= 0) {
      alert('Seats must be greater than 0');
      return;
    }
    
    if (this.selectedBus.fare_per_seat < 0) {
      alert('Fare must be 0 or greater');
      return;
    }

    // Ensure verification_documents is an array
    if (!this.selectedBus.verification_documents) {
      this.selectedBus.verification_documents = [];
    }

    // Create a clean update payload (exclude complex fields)
    const updateData = {
      id: this.selectedBus.id,
      bus_number: this.selectedBus.bus_number,
      company_name: this.selectedBus.company_name,
      identify_or_incorporation_no: this.selectedBus.identify_or_incorporation_no,
      business_email: this.selectedBus.business_email,
      business_phone: this.selectedBus.business_phone,
      permit_number: this.selectedBus.permit_number,
      license_plate: this.selectedBus.license_plate,
      total_seats: this.selectedBus.total_seats,
      bus_type: this.selectedBus.bus_type,
      custom_route_name: this.selectedBus.custom_route_name,
      fare_per_seat: this.selectedBus.fare_per_seat,
      status: this.selectedBus.status,
      verification_status: this.selectedBus.verification_status,
      verification_documents: this.selectedBus.verification_documents || []
    };

    console.log('Updating bus:', updateData);
    this.busService.updateBus(updateData as Bus).subscribe({
      next: (response) => {
        console.log('✓ Bus updated successfully:', response);
        alert('Successfully updated');
        this.busService.loadBuses(); // Reload buses to refresh the list
        this.closeEditBusModal();
      },
      error: (err) => {
        console.error('✗ Failed to update bus:', err);
        let errorMessage = 'Unknown error';
        if (err.error?.error) {
          const errorText = err.error.error;
          if (errorText.includes('buses_license_plate_key')) {
            errorMessage = `License plate "${this.selectedBus?.license_plate}" already exists. Please use a different registration number.`;
          } else if (errorText.includes('buses_bus_number_key')) {
            errorMessage = `Bus number "${this.selectedBus?.bus_number}" already exists. Please use a different bus number.`;
          } else if (errorText.includes('buses_permit_number_key')) {
            errorMessage = `Permit number "${this.selectedBus?.permit_number}" already exists. Please use a different permit number.`;
          } else if (errorText.includes('duplicate key')) {
            errorMessage = 'This bus details already exist in the system. Please check your input.';
          } else {
            errorMessage = errorText;
          }
        } else if (err.message) {
          errorMessage = err.message;
        }
        alert(`Failed to update bus: ${errorMessage}`);
      }
    });
  }

  updateBus(bus: Bus) {
    this.selectedBus = { ...bus };
    this.showEditBusModal = true;
  }

  toggleActive(bus: Bus) {
    const currentStatus = bus.status.toLowerCase();
    bus.status = currentStatus === 'active' ? 'inactive' : 'active';
    console.log(`${bus.bus_number} is now ${bus.status}`);
    // Call backend API to update status
    this.busService.updateBus(bus).subscribe({
      next: () => {
        this.busService.loadBuses(); // Reload buses to refresh the list
      },
      error: (err) => {
        console.error('Error updating status', err);
        // Revert status on error
        const revertStatus = bus.status.toLowerCase();
        bus.status = revertStatus === 'active' ? 'inactive' : 'active';
      }
    });
  }

  deleteBus(bus: Bus) {
    const confirmed = confirm(`Are you sure you want to delete ${bus.bus_number}?`);
    if (confirmed) {
      this.busService.deleteBus(bus.id).subscribe({
        next: () => {
          console.log(`${bus.bus_number} deleted`);
          this.busService.loadBuses(); // Reload buses to refresh the list
        },
        error: (err) => console.error('Error deleting bus', err)
      });
    }
  }

  exportHistoryPdf() {
    const doc = new jsPDF({ orientation: 'landscape' });
    doc.setFontSize(16);
    doc.text('Buses History', 14, 16);

    const tableHead = [['Bus ID', 'Company', 'Contact', 'Permit Num', 'Reg Num', 'Bus Type', 'Seats', 'Route', 'Approved Fare', 'Status', 'Verification Status']];
    const tableBody = this.buses.map(b => [
      b.id,
      b.company_name,
      b.business_phone,
      b.permit_number,
      b.license_plate,
      b.bus_type,
      String(b.total_seats),
      b.custom_route_name,
      String(b.fare_per_seat),
      b.status,
      b.verification_status
    ]);

    autoTable(doc, {
      head: tableHead,
      body: tableBody,
      startY: 22,
      styles: { fontSize: 8 },
      headStyles: { fillColor: [59, 130, 246] }
    });

    doc.save('buses-history.pdf');
  }

  // Stats helpers to keep template simple (no arrow functions in template)
  getTotalBuses(): number {
    return this.buses.length;
  }

  getActiveCount(): number {
    return this.buses.filter(bus => bus.status.toLowerCase() === 'active').length;
  }

  getInactiveCount(): number {
    return this.buses.filter(bus => bus.status.toLowerCase() === 'inactive').length;
  }

  getAverageCapacity(): number {
    if (this.buses.length === 0) return 0;
    const total = this.buses.reduce((sum, bus) => sum + bus.total_seats, 0);
    return Math.round(total / this.buses.length);
  }

  getTotalCapacity(): number {
    return this.buses.reduce((sum, bus) => sum + bus.total_seats, 0);
  }

  getBusCountByType(type: string): number {
    return this.buses.filter(bus => bus.bus_type === type).length;
  }

  getBusCountByTypePercentage(type: string): number {
    const total = this.buses.length || 1;
    return (this.getBusCountByType(type) / total) * 100;
  }

  getActivePercentage(): number {
    const total = this.buses.length || 1;
    return (this.getActiveCount() / total) * 100;
  }

  getInactivePercentage(): number {
    const total = this.buses.length || 1;
    return (this.getInactiveCount() / total) * 100;
  }

  getBusPieBackground(): string {
    const active = this.getActivePercentage();
    return `conic-gradient(#0046FF 0% ${active}%, #FAA533 ${active}% 100%)`;
  }

  updateBarChart(): void {
    this.barChartData = {
      labels: ['Normal', 'Semi-Luxury', 'Luxury'],
      datasets: [{
        data: [
          this.getBusCountByType('Normal'),
          this.getBusCountByType('Semi-Luxury'),
          this.getBusCountByType('Luxury')
        ],
        backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533'],
        borderColor: ['#0046FF', '#a3a3a3', '#FAA533'],
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

  // Search functionality
  onSearchChange(): void {
    this.applyFilters();
  }

  clearSearch(): void {
    this.searchTerm = '';
    this.applyFilters();
  }

  // Filter functionality
  onFilterChange(): void {
    this.applyFilters();
  }

  private applyFilters(): void {
    let filtered = this.buses;

    // Apply search filter based on selected criteria
    if (this.searchTerm.trim()) {
      const lowercasedTerm = this.searchTerm.toLowerCase();
      filtered = filtered.filter(bus => {
        const value = bus[this.searchCriteria as keyof Bus];
        if (value) {
          return value.toString().toLowerCase().includes(lowercasedTerm);
        }
        return false;
      });
    }

    // Apply status filter
    if (this.statusFilter !== 'all') {
      filtered = filtered.filter(bus => bus.status.toLowerCase() === this.statusFilter.toLowerCase());
    }

    this.filteredBuses = filtered;

    // Apply current sorting to filtered results
    if (this.sortColumn) {
      this.applySorting();
    }
  }

  goDriverManagement(): void {
    this.router.navigate(['/driver-management']);
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

  onNavigate(page: string): void { this.router.navigate([`/${page}`]); }
  
  logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('admin_user');
    this.router.navigate(['/login']);
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
    this.filteredBuses = [...this.filteredBuses].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (this.sortColumn) {
        case 'id':
          aValue = a.id;
          bValue = b.id;
          break;
        case 'company_name':
          aValue = a.company_name.toLowerCase();
          bValue = b.company_name.toLowerCase();
          break;
        case 'business_phone':
          aValue = a.business_phone.toLowerCase();
          bValue = b.business_phone.toLowerCase();
          break;
        case 'permit_number':
          aValue = a.permit_number.toLowerCase();
          bValue = b.permit_number.toLowerCase();
          break;
        case 'license_plate':
          aValue = a.license_plate.toLowerCase();
          bValue = b.license_plate.toLowerCase();
          break;
        case 'bus_type':
          aValue = a.bus_type.toLowerCase();
          bValue = b.bus_type.toLowerCase();
          break;
        case 'total_seats':
          aValue = a.total_seats;
          bValue = b.total_seats;
          break;
        case 'custom_route_name':
          aValue = a.custom_route_name || '';
          bValue = b.custom_route_name || '';
          break;
        case 'fare_per_seat':
          aValue = a.fare_per_seat;
          bValue = b.fare_per_seat;
          break;
        case 'status':
          aValue = a.status;
          bValue = b.status;
          break;
        case 'verification_status':
          aValue = a.verification_status.toLowerCase();
          bValue = b.verification_status.toLowerCase();
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

  // Getters and setters for documents field to handle string/array conversion
  get newBusDocuments(): string {
    return this.newBus.verification_documents ? this.newBus.verification_documents.join(', ') : '';
  }

  set newBusDocuments(value: string) {
    this.newBus.verification_documents = value ? value.split(',').map(s => s.trim()).filter(s => s) : [];
  }

  get selectedBusDocuments(): string {
    return this.selectedBus?.verification_documents ? this.selectedBus.verification_documents.join(', ') : '';
  }

  set selectedBusDocuments(value: string) {
    if (this.selectedBus) {
      this.selectedBus.verification_documents = value ? value.split(',').map(s => s.trim()).filter(s => s) : [];
    }
  }

  // Document upload modal methods
  closeDocumentModal() {
    this.showDocumentModal = false;
    this.selectedBusForDocuments = null;
    this.selectedFiles = [];
    this.selectedDocumentType = '';
  }

  uploadDocuments() {
    if (this.selectedFiles.length > 0 && this.selectedDocumentType && this.selectedBusForDocuments) {
      // TODO: Implement actual file upload logic
      console.log('Uploading documents for bus:', this.selectedBusForDocuments.bus_number);
      console.log('Document type:', this.selectedDocumentType);
      console.log('Files:', this.selectedFiles);
      this.closeDocumentModal();
    }
  }

  onFileSelected(event: any) {
    this.selectedFiles = Array.from(event.target.files);
  }

  removeFile(index: number) {
    this.selectedFiles.splice(index, 1);
  }
}
