import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { BaseChartDirective } from 'ng2-charts';
import { Chart, registerables } from 'chart.js';
import { Lounge } from '../../core/models/lounge.model';
import { LoungeService } from '../../core/services/lounge.service';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-lounges-management',
  standalone: true,
  imports: [CommonModule, FormsModule, BaseChartDirective, NotificationPanelComponent, NavbarComponent, RouterModule],
  templateUrl: './lounges-management.component.html',
  styleUrls: ['./lounges-management.component.scss']
})
export class LoungesManagementComponent implements OnInit {
  lounges: Lounge[] = [];
  filteredLounges: Lounge[] = [];
  searchTerm = '';
  priceFilter: number | null = null;
  showNotificationPanel = false;

  amenitiesCounts: { label: string; count: number }[] = [];
  servicesCounts: { label: string; count: number }[] = [];

  currentPage = 'lounges';

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // Modal properties
  showModal: boolean = false;
  modalMode: 'add' | 'view' | 'edit' = 'add';
  selectedLounge: Lounge | null = null;
  lounge: Lounge = {
    lounge_id: '0',
    lounge_owner: '',
    lounge_name: '',
    lounge_contact: '',
    address: '',
    capacity: 0,
    price_per_hour: 0,
    facilities: [],
    marketplace: '',
    verification: 'pending',
    verification_note: '',
    operational: true
  };

  selectedAmenities: string[] = [];
  selectedMarketplaceItems: string[] = [];

  // Updated to match Supabase DB values (snake_case)
  availableAmenities: string[] = ['wifi', 'waiting_area', 'ac', 'cafeteria', 'charging_ports', 'parking', 'restrooms', 'tv', 'quiet_zone'];
  availableServices: string[] = ['Food', 'Drinks', 'Essentials', 'Other'];

  imagePreviews: string[] = [];
  private selectedFiles: File[] = [];

  // Chart properties
  barChartData: any[] = [];
  barChartOptions: any;
  isBrowser: boolean;

  exportLoungeHistoryPdf(): void {
    const doc = new jsPDF({ orientation: 'landscape' });
    doc.setFontSize(16);
    doc.text('Lounges History', 14, 16);

    const tableHead = [['Lounge Name', 'Owner', 'Address', 'Lounge Contact', 'Capacity', 'Price/hr', 'Operational']];
    const tableBody = this.filteredLounges?.map((l: Lounge) => [
      l.lounge_name,
      l.lounge_owner,
      l.address,
      l.lounge_contact,
      String(l.capacity),
      `$${l.price_per_hour}`,
      l.operational ? 'Open' : 'Closed'
    ]) ?? [];

    autoTable(doc, {
      head: tableHead,
      body: tableBody,
      startY: 22,
      styles: { fontSize: 10 },
      headStyles: { fillColor: [59, 130, 246] }
    });

    doc.save('lounges-history.pdf');
  }

  constructor(private router: Router, private loungeService: LoungeService, @Inject(PLATFORM_ID) private platformId: Object, public notificationService: NotificationService) {
    this.isBrowser = isPlatformBrowser(this.platformId);
  }

  ngOnInit(): void {
    this.currentPage = 'lounges-management'; // Set currentPage to match sidebar item key
    if (this.isBrowser) {
      Chart.register(...registerables);
    }
    this.loungeService.lounges$.subscribe(ls => {
      this.lounges = ls;
      this.filteredLounges = ls;
      this.refreshCharts();
      if (this.isBrowser) {
        this.updateBarCharts();
      }
    });
  }

  // New method to get data for Capacity vs Price chart
  getCapacityPriceData() {
    return this.lounges.map(lounge => ({
      name: lounge.lounge_name,
      capacity: lounge.capacity,
      price_per_hour: lounge.price_per_hour
    }));
  }

  // New method to get data for Food, Drinks, Shower chart
  getFoodDrinksShowerData() {
    const servicesCounts = this.loungeService.getServicesCounts();
    const filtered = Object.entries(servicesCounts).filter(([label]) =>
      ['Food', 'Drinks', 'Shower', 'Essentials', 'Other'].includes(label)
    ).map(([label, count]) => ({ label, count }));
    return filtered;
  }

  // New method to get data for fixed Amenities chart (WiFi, AC, TV, Charging Ports, Quiet Zone)
  getFixedAmenitiesData() {
    // Use DB keys here
    const fixedAmenities = ['wifi', 'ac', 'tv', 'charging_ports', 'quiet_zone'];
    const amenitiesCounts = this.loungeService.getAmenitiesCounts();
    const filtered = fixedAmenities.map(key => ({
      label: this.formatAmenity(key), // Display nice name
      count: amenitiesCounts[key] || 0
    }));
    return filtered;
  }

  formatAmenity(amenity: string): string {
    if (!amenity) return '';
    // Special cases
    if (amenity === 'wifi') return 'WiFi';
    if (amenity === 'ac') return 'AC';
    if (amenity === 'tv') return 'TV';
    
    // General snake_case to Title Case
    return amenity
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ');
  }

  // New method to get data for Amenities & Services coverage chart
  // Removed duplicate getAmenitiesServicesData method to fix duplicate function implementation error

  navigateTo(page: string): void { this.router.navigate([`/${page}`]); }

  goDashboard(): void {
    this.currentPage = 'dashboard';
    this.router.navigate(['/dashboard']);
  }

  onSearchChange(): void {
    const q = this.searchTerm.toLowerCase();
    this.filteredLounges = this.lounges.filter(l => {
      const matchesSearch = !q ||
        l.lounge_owner.toLowerCase().includes(q) ||
        l.lounge_name.toLowerCase().includes(q) ||
        l.address.toLowerCase().includes(q) ||
        l.lounge_contact.toLowerCase().includes(q) ||
        l.lounge_id.toLowerCase().includes(q) ||
        l.capacity.toString().includes(q) ||
        l.price_per_hour.toString().includes(q) ||
        (l.facilities || []).join(' ').toLowerCase().includes(q) ||
        l.marketplace.toLowerCase().includes(q) ||
        l.verification.toLowerCase().includes(q) ||
        (l.verification_note?.toLowerCase() || '').includes(q) ||
        (l.operational ? 'open' : 'closed').includes(q);
      const matchesPrice = this.priceFilter === null || l.price_per_hour === this.priceFilter;
      return matchesSearch && matchesPrice;
    });
  }

  clearSearch(): void { this.searchTerm = ''; this.priceFilter = null; this.filteredLounges = this.lounges; }

  addLounge(): void {
    this.modalMode = 'add';
    this.resetAddLoungeForm();
    this.showModal = true;
  }

  private resetAddLoungeForm(): void {
    this.lounge = {
      lounge_id: '0',
      lounge_owner: '',
      lounge_name: '',
      lounge_contact: '',
      address: '',
      capacity: 0,
      price_per_hour: 0,
      facilities: [],
      marketplace: '',
      verification: 'pending',
      verification_note: '',
      operational: true
    };
    this.selectedAmenities = [];
    this.selectedMarketplaceItems = [];
    this.imagePreviews = [];
    this.selectedFiles = [];
  }

  // Validate lounge form - all required fields must be filled
  isLoungeFormValid(): boolean {
    return !!(
      this.lounge.lounge_owner?.trim() &&
      this.lounge.owner_nic?.trim() &&
      this.lounge.owner_email?.trim() &&
      this.lounge.owner_contact?.trim() &&
      this.lounge.lounge_name?.trim() &&
      this.lounge.lounge_contact?.trim() &&
      this.lounge.address?.trim() &&
      this.lounge.capacity > 0 &&
      this.lounge.price_per_hour >= 0
    );
  }

  saveAddLounge(): void {
    // Validate all required fields
    if (!this.isLoungeFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }

    if (this.modalMode === 'add') {
      this.lounge.facilities = this.selectedAmenities;
      this.lounge.marketplace = this.selectedMarketplaceItems.join(', ');
      this.loungeService.add(this.lounge).subscribe({
        next: () => {
          console.log('✓ Lounge added successfully');
          // No need to call loadLounges() - the service already does this via tap operator
          this.showModal = false;
          alert('Lounge added successfully');
        },
        error: (err) => {
          console.error('✗ Failed to add lounge:', err);
          alert(`Failed to add lounge: ${err.error?.error || err.message || 'Unknown error'}`);
        }
      });
    } else if (this.modalMode === 'edit') {
      this.lounge.facilities = this.selectedAmenities;
      this.lounge.marketplace = this.selectedMarketplaceItems.join(', ');
      this.loungeService.update(this.lounge).subscribe({
        next: () => {
          console.log('✓ Lounge updated successfully');
          // No need to call loadLounges() - the service already does this via tap operator
          this.showModal = false;
          alert('Lounge updated successfully');
        },
        error: (err) => {
          console.error('✗ Failed to update lounge:', err);
          alert(`Failed to update lounge: ${err.error?.error || err.message || 'Unknown error'}`);
        }
      });
    }
  }

  cancelAddLounge(): void {
    this.showModal = false;
  }

  toggleAmenity(amenity: string): void {
    const index = this.selectedAmenities.indexOf(amenity);
    if (index > -1) {
      this.selectedAmenities.splice(index, 1);
    } else {
      this.selectedAmenities.push(amenity);
    }
  }

  toggleMarketplaceItem(item: string): void {
    const index = this.selectedMarketplaceItems.indexOf(item);
    if (index > -1) {
      this.selectedMarketplaceItems.splice(index, 1);
    } else {
      this.selectedMarketplaceItems.push(item);
    }
  }

  toggleService(service: string): void {
    // const index = this.selectedServices.indexOf(service);
    // if (index > -1) {
    //   this.selectedServices.splice(index, 1);
    // } else {
    //   this.selectedServices.push(service);
    // }
  }

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    const files = input.files;
    if (!files || files.length === 0) return;

  

    // Append to existing selections to allow multiple picks across interactions
    for (const file of Array.from(files)) {
      if (!file.type.startsWith('image/')) continue; // skip non-images
      this.selectedFiles.push(file);

      const reader = new FileReader();
      reader.onload = (e) => {
        const result = (e.target as FileReader).result as string;
        this.imagePreviews.push(result);
      };
      reader.readAsDataURL(file); // create base64 preview
    }

    // Clear the input to allow re-selecting the same files if needed
    input.value = '';
  }
  
  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }
  view(l: Lounge): void {
    this.modalMode = 'view';
    this.lounge = { ...l };
    this.selectedAmenities = [...l.facilities];
    // this.selectedServices = [...l.services];
    // this.imagePreviews = [...l.images];
    this.showModal = true;
  }

  update(l: Lounge): void {
    this.modalMode = 'edit';
    this.lounge = { ...l };
    this.selectedAmenities = [...l.facilities];
    // this.selectedServices = [...l.services];
    // this.imagePreviews = [...l.images];
    this.showModal = true;
  }

  delete(l: Lounge): void {
    const ok = confirm(`Delete ${l.lounge_name}?`);
    if (ok) {
      this.loungeService.delete(l.lounge_id).subscribe({
        next: () => {
          console.log(`${l.lounge_name} deleted`);
          this.loungeService.loadLounges(); // Reload lounges to refresh the list
        },
        error: (err) => console.error('Failed to delete lounge', err)
      });
    }
  }

  refreshCharts(): void {
    const aCounts = this.loungeService.getAmenitiesCounts();
    // const sCounts = this.loungeService.getServicesCounts();
    this.amenitiesCounts = Object.entries(aCounts).map(([label, count]) => ({ label, count }));
    // this.servicesCounts = Object.entries(sCounts).map(([label, count]) => ({ label, count }));
  }

  updateBarCharts(): void {
    this.barChartData = [
      {
        labels: this.getCapacityPriceData().map(d => d.name),
        datasets: [{
          data: this.getCapacityPriceData().map(d => d.capacity),
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
          borderWidth: 0.25
        }]
      },
      {
        labels: this.getFoodDrinksShowerData().map(d => d.label),
        datasets: [{
          data: this.getFoodDrinksShowerData().map(d => d.count),
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533'],
          borderWidth: 0.25
        }]
      },
      {
        labels: this.getFixedAmenitiesData().map(d => d.label),
        datasets: [{
          data: this.getFixedAmenitiesData().map(d => d.count),
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#6db9f8ff', '#fad577'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#6db9f8ff', '#fad577'],
          borderWidth: 0.25
        }]
      }
    ];
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

  getAmenitiesServicesData() {
    return {
      amenities: this.amenitiesCounts,
      services: this.servicesCounts
    };
  }

  totalLounges(): number { return this.lounges.length; }
  showProfileMenu = false;

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

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
    this.filteredLounges = [...this.filteredLounges].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (this.sortColumn) {
        case 'lounge_id':
          aValue = a.lounge_id.toLowerCase();
          bValue = b.lounge_id.toLowerCase();
          break;
        case 'lounge_owner':
          aValue = a.lounge_owner.toLowerCase();
          bValue = b.lounge_owner.toLowerCase();
          break;
        case 'lounge_name':
          aValue = a.lounge_name.toLowerCase();
          bValue = b.lounge_name.toLowerCase();
          break;
        case 'lounge_contact':
          aValue = a.lounge_contact.toLowerCase();
          bValue = b.lounge_contact.toLowerCase();
          break;
        case 'address':
          aValue = a.address.toLowerCase();
          bValue = b.address.toLowerCase();
          break;
        case 'capacity':
          aValue = a.capacity;
          bValue = b.capacity;
          break;
        case 'price_per_hour':
          aValue = a.price_per_hour;
          bValue = b.price_per_hour;
          break;
        case 'verification':
          aValue = a.verification.toLowerCase();
          bValue = b.verification.toLowerCase();
          break;
        case 'verification_note':
          aValue = a.verification_note.toLowerCase();
          bValue = b.verification_note.toLowerCase();
          break;
        case 'operational':
          aValue = a.operational;
          bValue = b.operational;
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
}


