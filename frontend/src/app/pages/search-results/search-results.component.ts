import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { BusService } from '../../core/services/bus.service';
import { LoungeService } from '../../core/services/lounge.service';
import { DriverService } from '../../core/services/driver.service';
import { ConductorService } from '../../core/services/conductor.service';
import { BusBookingService } from '../../core/services/bus-booking.service';
import { LoungeBookingService } from '../../core/services/lounge-booking.service';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-search-results',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent],
  templateUrl: './search-results.component.html',
  styleUrls: ['./search-results.component.scss']
})
export class SearchResultsComponent implements OnInit {
  searchType: string = '';
  searchCriteria: any = {};
  searchResults: any[] = [];
  showProfileMenu = false;
  selectedAttributes: string[] = [];
  showAddModal = false;
  showEditModal = false;
  showViewModal = false;
  newEntity: any = {};
  selectedEntity: any = null;
  modalMode: 'add' | 'edit' | 'view' = 'add';
  entityType: 'bus' | 'driver' | 'conductor' | 'lounge' | 'bus-booking' | 'lounge-booking' = 'bus';
  newBusDocuments: string = '';
  editBusDocuments: string = '';
  formSeatNumbers: string = ''; // For bus booking seat numbers
  availableFeatures = ['Premium meals', 'Express loundary', 'cargo storage', 'spa service', 'personal assist', 'Airport transfer', 'Tuk tuk'];
  
  // Lounge amenities and services
  availableAmenities: string[] = ['wifi', 'waiting_area', 'ac', 'cafeteria', 'charging_ports', 'parking', 'restrooms', 'tv', 'quiet_zone'];
  availableServices: string[] = ['Food', 'Drinks', 'Essentials', 'Other'];
  selectedAmenities: string[] = [];
  selectedMarketplaceItems: string[] = [];

  attributeOptions: { [key: string]: string[] } = {
    Bus: ['Company', 'Route', 'Permit Num', 'Register Num', 'Owner Verification', 'Permit Verify', 'Contact', 'No of Seat', 'Approved fare', 'Type', 'Status'],
    Lounge: ['Lounge Name', 'Owner', 'Contact', 'Address', 'Price per hour', 'Capacity', 'Operation', 'Verification'],
    Driver: ['Name', 'Contact', 'License Num', 'License Expire date', 'Experience', 'Hire date', 'Verification', 'Status'],
    Conductor: ['Name', 'Contact', 'License Num', 'Experience', 'Hire date', 'Verification', 'Status'],
    'Lounge booking': ['Passenger Name', 'Passenger Phone', 'Ref NUM', 'Lounge Name', 'Market place', 'Booking Type', 'Date', 'Duration', 'No of Guests', 'Total Amount', 'Payment Status', 'Booking Status'],
    'Bus booking': ['Bus Number', 'Passenger Name', 'Passenger Phone', 'Ref NUM', 'Route', 'Date & Time', 'Bus Type', 'Seat No', 'Total Fare', 'Payment Status', 'Booking Status']
  };

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private busService: BusService,
    private loungeService: LoungeService,
    private driverService: DriverService,
    private conductorService: ConductorService,
    private busBookingService: BusBookingService,
    private loungeBookingService: LoungeBookingService,
    public notificationService: NotificationService
  ) {}

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      this.searchType = params['type'] || 'Bus';
      const criteriaStr = params['criteria'];
      if (criteriaStr) {
        this.searchCriteria = JSON.parse(criteriaStr);
        this.selectedAttributes = Object.keys(this.searchCriteria).filter(key => this.searchCriteria[key]);
      }
      // Always load results for the selected type
      this.performSearch();
    });
  }

  performSearch() {
    switch (this.searchType) {
      case 'Bus':
        this.searchBuses();
        break;
      case 'Lounge':
        this.searchLounges();
        break;
      case 'Driver':
        this.searchDrivers();
        break;
      case 'Conductor':
        this.searchConductors();
        break;
      case 'Bus booking':
        this.searchBusBookings();
        break;
      case 'Lounge booking':
        this.searchLoungeBookings();
        break;
    }
  }

  private searchBuses() {
    const buses = this.busService.buses;
    this.searchResults = buses.filter(bus => {
      return Object.keys(this.searchCriteria).every(attr => {
        const searchValue = this.searchCriteria[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Company':
            return bus.company_name?.toLowerCase().includes(searchValue);
          case 'Route':
            return bus.custom_route_name?.toLowerCase().includes(searchValue);
          case 'Permit Num':
            return bus.permit_number?.toLowerCase().includes(searchValue);
          case 'Register Num':
            return bus.license_plate?.toLowerCase().includes(searchValue);
          case 'Owner Verification':
            return bus.owner_verification_status?.toLowerCase().includes(searchValue);
          case 'Permit Verify':
            return bus.verification_status?.toLowerCase().includes(searchValue);
          case 'Contact':
            return bus.business_phone?.toLowerCase().includes(searchValue);
          case 'No of Seat':
            return bus.total_seats?.toString() === searchValue;
          case 'Approved fare':
            return bus.fare_per_seat?.toString() === searchValue;
          case 'Type':
            // Normalize both search value and bus type to match variations (semi-luxury, semi_luxury, etc.)
            const normalizedSearchValue = searchValue.replace(/[-_]/g, '');
            const normalizedBusType = bus.bus_type?.toLowerCase().replace(/[-_]/g, '');
            return normalizedBusType?.includes(normalizedSearchValue);
          case 'Status':
            return bus.status?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  private searchLounges() {
    const lounges = this.loungeService.lounges;
    this.searchResults = lounges.filter(lounge => {
      return Object.keys(this.searchCriteria).every(attr => {
        const searchValue = this.searchCriteria[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Lounge Name':
            return lounge.lounge_name?.toLowerCase().includes(searchValue);
          case 'Owner':
            return lounge.lounge_owner?.toLowerCase().includes(searchValue);
          case 'Contact':
            return lounge.lounge_contact?.toLowerCase().includes(searchValue);
          case 'Address':
            return lounge.address?.toLowerCase().includes(searchValue);
          case 'Price per hour':
            return lounge.price_per_hour?.toString() === searchValue;
          case 'Capacity':
            return lounge.capacity?.toString() === searchValue;
          case 'Operation':
            const status = typeof lounge.operational === 'string' ? lounge.operational : (lounge.operational ? 'open' : 'closed');
            return status.toLowerCase().includes(searchValue);
          case 'Verification':
            return lounge.verification?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  private searchDrivers() {
    const drivers = this.driverService.drivers;
    this.searchResults = drivers.filter(driver => {
      return Object.keys(this.searchCriteria).every(attr => {
        const searchValue = this.searchCriteria[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Name':
            return driver.name?.toLowerCase().includes(searchValue);
          case 'Contact':
            return driver.contact_number?.toLowerCase().includes(searchValue);
          case 'License Num':
            return driver.license_number?.toLowerCase().includes(searchValue);
          case 'License Expire date':
            return driver.license_expiry_date?.toLowerCase().includes(searchValue);
          case 'Experience':
            return driver.experience_years?.toString() === searchValue;
          case 'Hire date':
            return driver.hire_date?.toLowerCase().includes(searchValue);
          case 'Verification':
            return driver.verification_status?.toLowerCase().includes(searchValue);
          case 'Status':
            return driver.status?.toLowerCase() === searchValue;
          default:
            return true;
        }
      });
    });
  }

  private searchConductors() {
    const conductors = this.conductorService.conductors;
    this.searchResults = conductors.filter(conductor => {
      return Object.keys(this.searchCriteria).every(attr => {
        const searchValue = this.searchCriteria[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Name':
            return conductor.name?.toLowerCase().includes(searchValue);
          case 'Contact':
            return conductor.contact_number?.toLowerCase().includes(searchValue);
          case 'License Num':
            return conductor.license_number?.toLowerCase().includes(searchValue);
          case 'Experience':
            return conductor.experience_years?.toString() === searchValue;
          case 'Hire date':
            return conductor.hire_date?.toLowerCase().includes(searchValue);
          case 'Verification':
            return conductor.verification_status?.toLowerCase().includes(searchValue);
          case 'Status':
            return conductor.status?.toLowerCase() === searchValue;
          default:
            return true;
        }
      });
    });
  }

  private searchBusBookings() {
    const bookings = this.busBookingService.bookings;
    this.searchResults = bookings.filter(booking => {
      return Object.keys(this.searchCriteria).every(attr => {
        const searchValue = this.searchCriteria[attr]?.toLowerCase().trim();
        if (!searchValue) return true;

        switch (attr) {
          case 'Bus Number':
            return booking.bus_number?.toLowerCase().includes(searchValue);
          case 'Passenger Name':
            return booking.passenger_name?.toLowerCase().includes(searchValue);
          case 'Passenger Phone':
            return booking.passenger_phone?.toLowerCase().includes(searchValue);
          case 'Ref NUM':
            return booking.booking_reference?.toLowerCase().includes(searchValue);
          case 'Route':
            return booking.route?.toLowerCase().includes(searchValue);
          case 'Date & Time':
            return booking.departure_datetime?.toLowerCase().includes(searchValue);
          case 'Bus Type':
            // Normalize both search value and bus type to match variations (semi-luxury, semi_luxury, etc.)
            const normalizedSearchValue = searchValue.replace(/[-_]/g, '');
            const normalizedBusType = booking.bus_type?.toLowerCase().replace(/[-_]/g, '');
            return normalizedBusType?.includes(normalizedSearchValue);
          case 'Seat No':
            return booking.seat_number?.toLowerCase().includes(searchValue);
          case 'Total Fare':
            return booking.total_fare?.toString() === searchValue;
          case 'Payment Status':
            return booking.payment_status?.toLowerCase().includes(searchValue);
          case 'Booking Status':
            return booking.booking_status?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  private searchLoungeBookings() {
    const bookings = this.loungeBookingService.bookings;
    this.searchResults = bookings.filter(booking => {
      return Object.keys(this.searchCriteria).every(attr => {
        const searchValue = this.searchCriteria[attr]?.toLowerCase();
        if (!searchValue) return true;

        switch (attr) {
          case 'Passenger Name':
            return booking.passenger_name?.toLowerCase().includes(searchValue);
          case 'Passenger Phone':
            return booking.passenger_phone?.toLowerCase().includes(searchValue);
          case 'Ref NUM':
            return booking.booking_reference?.toLowerCase().includes(searchValue);
          case 'Lounge Name':
            return booking.lounge_name?.toLowerCase().includes(searchValue);
          case 'Market place':
            return booking.product_name?.toLowerCase().includes(searchValue);
          case 'Booking Type':
            return booking.booking_type?.toLowerCase().includes(searchValue);
          case 'Date':
            // Extract date only (YYYY-MM-DD format) from scheduled_arrival for comparison
            const bookingDate = booking.scheduled_arrival ? booking.scheduled_arrival.split('T')[0] : '';
            return bookingDate.toLowerCase().includes(searchValue);
          case 'Duration':
            return booking.pricing_type?.toLowerCase().includes(searchValue);
          case 'No of Guests':
            return booking.number_of_guests?.toString() === searchValue;
          case 'Total Amount':
            return booking.total_amount?.toString() === searchValue;
          case 'Payment Status':
            return booking.payment_status?.toLowerCase().includes(searchValue);
          case 'Booking Status':
            return booking.status?.toLowerCase().includes(searchValue);
          default:
            return true;
        }
      });
    });
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  logout() {
    this.router.navigate(['/login']);
  }

  backToDashboard() {
    this.router.navigate(['/dashboard']);
  }

  selectSearchType(type: string) {
    this.searchType = type;
    this.selectedAttributes = [];
    this.searchCriteria = {};
    this.searchResults = [];
    // Automatically load all results for the selected type
    this.performSearch();
  }

  toggleAttribute(attribute: string) {
    const index = this.selectedAttributes.indexOf(attribute);
    if (index > -1) {
      this.selectedAttributes.splice(index, 1);
      delete this.searchCriteria[attribute];
    } else {
      this.selectedAttributes.push(attribute);
      this.searchCriteria[attribute] = '';
    }
  }

  isAttributeSelected(attribute: string): boolean {
    return this.selectedAttributes.includes(attribute);
  }

  executeSearch() {
    if (this.selectedAttributes.length === 0) {
      alert('Please select at least one search attribute');
      return;
    }

    const hasValue = this.selectedAttributes.some(attr => this.searchCriteria[attr]?.trim());
    if (!hasValue) {
      alert('Please enter at least one search value');
      return;
    }

    this.performSearch();
  }

  clearSearch() {
    this.selectedAttributes = [];
    this.searchCriteria = {};
    this.searchResults = [];
  }

  // Lounge booking status change methods
  changePaymentStatus(b: any, v: 'pending'|'paid'|'failed') {
    this.loungeBookingService.updatePaymentStatus(b.lounge_booking_id, v).subscribe({
      next: () => {
        b.payment_status = v;
      },
      error: (err) => {
        console.error('Error updating payment status:', err);
      }
    });
  }
  
  changeBookingStatus(b: any, v: 'confirmed'|'pending'|'cancelled'|'completed') {
    this.loungeBookingService.updateBookingStatus(b.lounge_booking_id, v).subscribe({
      next: () => {
        b.status = v;
      },
      error: (err) => {
        console.error('Error updating booking status:', err);
      }
    });
  }

  openEditModal(b: any) {
    // Navigate to lounge booking page with edit modal
    this.router.navigate(['/lounge-booking'], { 
      queryParams: { 
        edit: b.lounge_booking_id 
      } 
    });
  }

  navigateToAdd() {
    this.initializeNewEntity();
    this.showAddModal = true;
  }

  editBus(bus: any) {
    this.router.navigate(['/bus-management'], {
      queryParams: { editBusId: bus.id }
    });
  }

  viewBus(bus: any) {
    this.selectedEntity = { ...bus };
    this.modalMode = 'view';
    this.entityType = 'bus';
    this.showViewModal = true;
  }

  editDriver(driver: any) {
    this.selectedEntity = { ...driver };
    this.modalMode = 'edit';
    this.entityType = 'driver';
    this.showEditModal = true;
  }

  editConductor(conductor: any) {
    this.selectedEntity = { ...conductor };
    this.modalMode = 'edit';
    this.entityType = 'conductor';
    this.showEditModal = true;
  }

  editLounge(lounge: any) {
    this.selectedEntity = { ...lounge };
    this.modalMode = 'edit';
    this.entityType = 'lounge';
    this.showEditModal = true;
  }

  viewLounge(lounge: any) {
    this.selectedEntity = { ...lounge };
    this.modalMode = 'view';
    this.entityType = 'lounge';
    this.showViewModal = true;
  }

  editBusBooking(booking: any) {
    this.selectedEntity = { ...booking };
    this.modalMode = 'edit';
    this.entityType = 'bus-booking';
    this.showEditModal = true;
  }

  editLoungeBooking(booking: any) {
    this.selectedEntity = { ...booking };
    if (!this.selectedEntity.selected_amenities) {
      this.selectedEntity.selected_amenities = [];
    }
    this.modalMode = 'edit';
    this.entityType = 'lounge-booking';
    this.showEditModal = true;
  }

  closeEditModal() {
    this.showEditModal = false;
    this.selectedEntity = null;
  }

  closeViewModal() {
    this.showViewModal = false;
    this.selectedEntity = null;
  }

  // Validation methods for forms
  isFormValid(): boolean {
    switch (this.searchType) {
      case 'Bus':
        return this.isBusFormValid();
      case 'Driver':
        return this.isDriverFormValid();
      case 'Conductor':
        return this.isConductorFormValid();
      case 'Lounge':
        return this.isLoungeFormValid();
      case 'Bus booking':
        return this.isBusBookingFormValid();
      case 'Lounge booking':
        return this.isLoungeBookingFormValid();
      default:
        return false;
    }
  }

  isBusFormValid(): boolean {
    const entity = this.modalMode === 'add' ? this.newEntity : this.selectedEntity;
    return !!(
      entity.bus_number?.trim() &&
      entity.company_name?.trim() &&
      entity.identify_or_incorporation_no?.trim() &&
      entity.business_email?.trim() &&
      entity.business_phone?.trim() &&
      entity.permit_number?.trim() &&
      entity.license_plate?.trim() &&
      entity.custom_route_name?.trim() &&
      entity.total_seats > 0 &&
      entity.fare_per_seat >= 0
    );
  }

  isDriverFormValid(): boolean {
    const entity = this.modalMode === 'add' ? this.newEntity : this.selectedEntity;
    return !!(
      entity.name?.trim() &&
      entity.contact_number?.trim() &&
      entity.license_number?.trim() &&
      entity.license_expiry_date &&
      entity.experience_years !== null &&
      entity.experience_years !== undefined &&
      entity.hire_date
    );
  }

  isConductorFormValid(): boolean {
    const entity = this.modalMode === 'add' ? this.newEntity : this.selectedEntity;
    return !!(
      entity.name?.trim() &&
      entity.contact_number?.trim() &&
      entity.license_number?.trim() &&
      entity.license_expiry_date &&
      entity.experience_years !== null &&
      entity.experience_years !== undefined &&
      entity.hire_date
    );
  }

  isLoungeFormValid(): boolean {
    const entity = this.modalMode === 'add' ? this.newEntity : this.selectedEntity;
    return !!(
      entity.lounge_owner?.trim() &&
      entity.owner_nic?.trim() &&
      entity.owner_email?.trim() &&
      entity.owner_contact?.trim() &&
      entity.lounge_name?.trim() &&
      entity.lounge_contact?.trim() &&
      entity.address?.trim() &&
      entity.capacity > 0 &&
      entity.price_per_hour >= 0
    );
  }

  isBusBookingFormValid(): boolean {
    const entity = this.modalMode === 'add' ? this.newEntity : this.selectedEntity;
    return !!(
      entity.passenger_name?.trim() &&
      entity.route?.trim() &&
      entity.departure_datetime &&
      this.formSeatNumbers?.trim() &&
      entity.total_fare !== null &&
      entity.total_fare !== undefined &&
      entity.total_fare >= 0 &&
      entity.bus_type
    );
  }

  isLoungeBookingFormValid(): boolean {
    const entity = this.modalMode === 'add' ? this.newEntity : this.selectedEntity;
    return !!(
      entity.passenger_name?.trim() &&
      entity.passenger_phone?.trim() &&
      entity.lounge_name?.trim() &&
      entity.scheduled_arrival &&
      entity.pricing_type?.trim() &&
      entity.number_of_guests !== null &&
      entity.number_of_guests !== undefined &&
      entity.number_of_guests > 0 &&
      entity.total_amount !== null &&
      entity.total_amount !== undefined &&
      entity.total_amount >= 0
    );
  }

  saveEdit() {
    // Validate all required fields
    if (!this.isFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }

    switch (this.searchType) {
      case 'Bus':
        if (this.editBusDocuments) {
          this.selectedEntity.verification_documents = this.editBusDocuments.split(',').map((doc: string) => doc.trim());
        }
        this.busService.updateBus(this.selectedEntity).subscribe({
          next: () => {
            alert('Bus updated successfully!');
            this.closeEditModal();
            setTimeout(() => this.performSearch(), 300);
          },
          error: (err: any) => {
            console.error('Error updating bus:', err);
            alert('Error updating bus: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Driver':
        this.driverService.updateDriver(this.selectedEntity).subscribe({
          next: () => {
            alert('Driver updated successfully!');
            this.closeEditModal();
            setTimeout(() => this.performSearch(), 300);
          },
          error: (err: any) => {
            console.error('Error updating driver:', err);
            alert('Error updating driver: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Conductor':
        this.conductorService.updateConductor(this.selectedEntity).subscribe({
          next: () => {
            alert('Conductor updated successfully!');
            this.closeEditModal();
            setTimeout(() => this.performSearch(), 300);
          },
          error: (err: any) => {
            console.error('Error updating conductor:', err);
            alert('Error updating conductor: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Lounge':
        this.loungeService.update(this.selectedEntity).subscribe({
          next: () => {
            alert('Lounge updated successfully!');
            this.closeEditModal();
            setTimeout(() => this.performSearch(), 300);
          },
          error: (err: any) => {
            console.error('Error updating lounge:', err);
            alert('Error updating lounge: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Bus booking':
        this.busBookingService.update(this.selectedEntity).subscribe({
          next: () => {
            alert('Bus booking updated successfully!');
            this.closeEditModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err: any) => {
            console.error('Error updating bus booking:', err);
            alert('Error updating bus booking: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Lounge booking':
        this.loungeBookingService.update(this.selectedEntity).subscribe({
          next: () => {
            alert('Lounge booking updated successfully!');
            this.closeEditModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err: any) => {
            console.error('Error updating lounge booking:', err);
            alert('Error updating lounge booking: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
    }
  }

  initializeNewEntity() {
    switch (this.searchType) {
      case 'Bus':
        this.newEntity = {
          company_name: '',
          identify_or_incorporation_no: '',
          business_email: '',
          business_phone: '',
          bus_number: '',
          permit_number: '',
          license_plate: '',
          total_seats: 0,
          bus_type: 'Standard',
          custom_route_name: '',
          fare_per_seat: 0,
          status: 'inactive',
          verification_status: 'Pending',
          documents: ''
        };
        break;
      case 'Driver':
        this.newEntity = {
          name: '',
          contact_number: '',
          license_number: '',
          experience_years: 0,
          license_expiry_date: '',
          hire_date: '',
          verification_status: 'pending',
          verification_notes: '',
          status: 'inactive'
        };
        break;
      case 'Conductor':
        this.newEntity = {
          name: '',
          contact_number: '',
          license_number: '',
          experience_years: 0,
          license_expiry_date: '',
          hire_date: '',
          verification_status: 'pending',
          verification_notes: '',
          status: 'inactive'
        };
        break;
      case 'Lounge':
        this.newEntity = {
          lounge_owner: '',
          owner_nic: '',
          owner_email: '',
          owner_contact: '',
          lounge_name: '',
          lounge_contact: '',
          address: '',
          price_per_hour: 0,
          capacity: 0,
          facilities: [],
          marketplace: '',
          verification: 'pending',
          verification_note: '',
          operational: true
        };
        this.selectedAmenities = [];
        this.selectedMarketplaceItems = [];
        break;
      case 'Bus booking':
        this.newEntity = {
          booking_id: `BBK-${Math.floor(Math.random() * 10000)}`,
          passenger_name: '',
          passenger_phone: '',
          booking_reference: '',
          scheduled_trip_id: '',
          bus_id: '',
          bus_number: '',
          route: '',
          departure_datetime: '',
          bus_type: 'normal',
          seat_number: '',
          total_fare: 0,
          payment_status: 'pending',
          booking_status: 'pending',
          license_plate: '',
          number_of_seats: 0,
          created_at: new Date().toISOString()
        };
        this.formSeatNumbers = '';
        break;
      case 'Lounge booking':
        this.newEntity = {
          passenger_name: '',
          passenger_phone: '',
          booking_reference: '',
          lounge_name: '',
          product_name: '',
          booking_type: '',
          scheduled_arrival: '',
          pricing_type: '',
          number_of_guests: 1,
          total_amount: 0,
          selected_amenities: [],
          payment_status: 'pending',
          status: 'pending'
        };
        break;
    }
  }

  closeAddModal() {
    this.showAddModal = false;
    this.newEntity = {};
    this.selectedAmenities = [];
    this.selectedMarketplaceItems = [];
    this.formSeatNumbers = '';
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
  
  formatAmenity(amenity: string): string {
    if (!amenity) return '';
    if (amenity === 'wifi') return 'WiFi';
    if (amenity === 'ac') return 'AC';
    if (amenity === 'tv') return 'TV';
    return amenity.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ');
  }

  isFeatureSelected(feature: string): boolean {
    return this.newEntity.selected_amenities?.includes(feature) || false;
  }

  toggleFeature(feature: string) {
    if (!this.newEntity.selected_amenities) {
      this.newEntity.selected_amenities = [];
    }
    const index = this.newEntity.selected_amenities.indexOf(feature);
    if (index > -1) {
      this.newEntity.selected_amenities.splice(index, 1);
    } else {
      this.newEntity.selected_amenities.push(feature);
    }
  }

  saveNewEntity() {
    // Validate all required fields
    if (!this.isFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }

    switch (this.searchType) {
      case 'Bus':
        // Transform documents string to verification_documents array
        const busData = {
          ...this.newEntity,
          verification_documents: this.newEntity.documents 
            ? this.newEntity.documents.split(',').map((doc: string) => doc.trim()).filter((doc: string) => doc)
            : []
        };
        delete busData.documents;
        
        this.busService.addBus(busData).subscribe({
          next: () => {
            alert('Bus added successfully!');
            this.closeAddModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err) => {
            console.error('Error adding bus:', err);
            alert('Failed to add bus: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Driver':
        this.driverService.addDriver(this.newEntity).subscribe({
          next: () => {
            alert('Driver added successfully!');
            this.closeAddModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err) => {
            console.error('Error adding driver:', err);
            alert('Failed to add driver: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Conductor':
        this.conductorService.addConductor(this.newEntity).subscribe({
          next: () => {
            alert('Conductor added successfully!');
            this.closeAddModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err) => {
            console.error('Error adding conductor:', err);
            alert('Failed to add conductor: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Lounge':
        // Set facilities and marketplace from selections
        this.newEntity.facilities = this.selectedAmenities;
        this.newEntity.marketplace = this.selectedMarketplaceItems.join(', ');
        
        this.loungeService.add(this.newEntity).subscribe({
          next: () => {
            alert('Lounge added successfully!');
            this.closeAddModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err: any) => {
            console.error('Error adding lounge:', err);
            alert('Failed to add lounge: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Bus booking':
        // Set seat_number from formSeatNumbers and calculate number_of_seats
        this.newEntity.seat_number = this.formSeatNumbers;
        const seatCount = this.formSeatNumbers.split(',').filter((s: string) => s.trim()).length;
        this.newEntity.number_of_seats = seatCount;
        
        this.busBookingService.add(this.newEntity).subscribe({
          next: () => {
            alert('Bus booking added successfully!');
            this.closeAddModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err: any) => {
            console.error('Error adding bus booking:', err);
            alert('Failed to add bus booking: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
      case 'Lounge booking':
        this.loungeBookingService.add(this.newEntity).subscribe({
          next: () => {
            alert('Lounge booking added successfully!');
            this.closeAddModal();
            setTimeout(() => this.performSearch(), 500);
          },
          error: (err: any) => {
            console.error('Error adding lounge booking:', err);
            alert('Failed to add lounge booking: ' + (err.error?.error || err.message || 'Unknown error'));
          }
        });
        break;
    }
  }
}
