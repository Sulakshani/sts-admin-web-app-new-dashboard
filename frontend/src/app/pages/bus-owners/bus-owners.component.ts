import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { HttpClient } from '@angular/common/http';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NotificationService } from '../../core/services/notification.service';
import { environment } from '../../../environments/environment';

interface BusOwner {
  id: string;
  user_id: string;
  company_name: string;
  business_email: string;
  business_phone: string;
  identity_or_incorporation_no: string;
  verification_status: string;
  verification_documents: any;
  license_number?: string;
  contact_person?: string;
  address?: string;
  city?: string;
  state?: string;
  country?: string;
  postal_code?: string;
  tax_id?: string;
  total_buses?: number;
  profile_completed?: boolean;
  created_at?: string;
  updated_at?: string;
}

@Component({
  selector: 'app-bus-owners',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent, NotificationPanelComponent, RouterModule],
  templateUrl: './bus-owners.component.html',
  styleUrls: ['./bus-owners.component.scss']
})
export class BusOwnersComponent implements OnInit {
  busOwners: BusOwner[] = [];
  filteredBusOwners: BusOwner[] = [];
  searchTerm: string = '';
  showNotificationPanel = false;
  showProfileMenu = false;
  showAddOwnerModal = false;
  isEditMode = false;
  editingOwnerId: string = '';

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // New bus owner form
  newOwner = {
    company_name: '',
    business_email: '',
    business_phone: '',
    identity_or_incorporation_no: '',
    user_id: ''
  };

  private apiUrl = `${environment.apiUrl}/bus-owners`;

  constructor(
    private router: Router,
    private http: HttpClient,
    public notificationService: NotificationService
  ) {}

  ngOnInit(): void {
    this.loadBusOwners();
  }

  loadBusOwners(): void {
    this.http.get<BusOwner[]>(this.apiUrl).subscribe({
      next: (data) => {
        this.busOwners = data;
        this.filteredBusOwners = data;
      },
      error: (error) => {
        console.error('Error loading bus owners:', error);
        this.busOwners = [];
        this.filteredBusOwners = [];
      }
    });
  }

  getTotalBusOwners(): number {
    return this.busOwners.length;
  }

  getPendingVerifications(): number {
    return this.busOwners.filter(owner => owner.verification_status?.toLowerCase() === 'pending').length;
  }

  getVerifiedCount(): number {
    return this.busOwners.filter(owner => owner.verification_status?.toLowerCase() === 'verified').length;
  }

  onSearchChange(): void {
    this.applyFilters();
  }

  applyFilters(): void {
    this.filteredBusOwners = this.busOwners.filter(owner => {
      const searchLower = this.searchTerm.toLowerCase();
      return (
        owner.id.toString().includes(searchLower) ||
        (owner.company_name && owner.company_name.toLowerCase().includes(searchLower)) ||
        (owner.business_email && owner.business_email.toLowerCase().includes(searchLower)) ||
        (owner.business_phone && owner.business_phone.toLowerCase().includes(searchLower)) ||
        (owner.identity_or_incorporation_no && owner.identity_or_incorporation_no.toLowerCase().includes(searchLower)) ||
        owner.verification_status.toLowerCase().includes(searchLower)
      );
    });
  }

  clearSearch(): void {
    this.searchTerm = '';
    this.applyFilters();
  }

  onSort(column: string): void {
    if (this.sortColumn === column) {
      this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortColumn = column;
      this.sortDirection = 'asc';
    }

    this.filteredBusOwners.sort((a, b) => {
      let aValue: any = a[column as keyof BusOwner];
      let bValue: any = b[column as keyof BusOwner];

      if (typeof aValue === 'string') {
        aValue = aValue.toLowerCase();
        bValue = bValue.toLowerCase();
      }

      if (aValue < bValue) return this.sortDirection === 'asc' ? -1 : 1;
      if (aValue > bValue) return this.sortDirection === 'asc' ? 1 : -1;
      return 0;
    });
  }

  getSortIcon(column: string): string {
    if (this.sortColumn !== column) return '⇅';
    return this.sortDirection === 'asc' ? '↑' : '↓';
  }

  getVerificationBadgeClass(status: string): string {
    switch (status) {
      case 'Verified':
        return 'badge-verified';
      case 'Pending':
        return 'badge-pending';
      case 'Rejected':
        return 'badge-rejected';
      default:
        return '';
    }
  }

  addOwner(): void {
    this.showAddOwnerModal = true;
  }

  closeAddOwnerModal(): void {
    this.showAddOwnerModal = false;
    this.resetForm();
  }

  resetForm(): void {
    this.newOwner = {
      company_name: '',
      business_email: '',
      business_phone: '',
      identity_or_incorporation_no: '',
      user_id: ''
    };
  }

  saveOwner(): void {
    // Validate required fields
    if (!this.newOwner.company_name || !this.newOwner.business_email || 
        !this.newOwner.business_phone || !this.newOwner.identity_or_incorporation_no) {
      alert('Please fill all required fields');
      return;
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(this.newOwner.business_email)) {
      alert('Please enter a valid email address');
      return;
    }

    if (this.isEditMode) {
      // Update existing owner
      this.http.put<BusOwner>(`${this.apiUrl}/${this.editingOwnerId}`, this.newOwner).subscribe({
        next: (data) => {
          const index = this.busOwners.findIndex(o => o.id === this.editingOwnerId);
          if (index > -1) {
            this.busOwners[index] = data;
            this.applyFilters();
          }
          alert('Bus owner updated successfully!');
          this.closeAddOwnerModal();
        },
        error: (error) => {
          console.error('Error updating bus owner:', error);
          alert('Failed to update bus owner. Please try again.');
        }
      });
    } else {
      // Create new owner via API
      this.http.post<BusOwner>(this.apiUrl, this.newOwner).subscribe({
        next: (data) => {
          this.busOwners.push(data);
          this.applyFilters();
          alert('Bus owner added successfully!');
          this.closeAddOwnerModal();
        },
        error: (error) => {
          console.error('Error creating bus owner:', error);
          alert('Failed to add bus owner. Please try again.');
        }
      });
    }
  }

  viewDocuments(owner: BusOwner): void {
    console.log('View documents for:', owner);
    // Implement document viewing logic
  }

  editOwner(owner: BusOwner): void {
    this.isEditMode = true;
    this.editingOwnerId = owner.id;
    this.newOwner = {
      company_name: owner.company_name,
      business_email: owner.business_email,
      business_phone: owner.business_phone,
      identity_or_incorporation_no: owner.identity_or_incorporation_no,
      user_id: owner.user_id
    };
    this.showAddOwnerModal = true;
  }

  deleteOwner(owner: BusOwner): void {
    if (confirm(`Are you sure you want to delete ${owner.company_name}?`)) {
      this.http.delete(`${this.apiUrl}/${owner.id}`).subscribe({
        next: () => {
          const index = this.busOwners.findIndex(o => o.id === owner.id);
          if (index > -1) {
            this.busOwners.splice(index, 1);
            this.applyFilters();
          }
        },
        error: (error) => {
          console.error('Error deleting bus owner:', error);
          alert('Failed to delete bus owner. Please try again.');
        }
      });
    }
  }

  goDashboard(): void {
    this.router.navigate(['/dashboard']);
  }

  toggleProfileMenu(): void {
    this.showProfileMenu = !this.showProfileMenu;
  }

  toggleNotificationPanel(): void {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel(): void {
    this.showNotificationPanel = false;
  }

  logout(): void {
    this.router.navigate(['/login']);
  }
}
