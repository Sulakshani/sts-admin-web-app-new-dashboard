import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NotificationService } from '../../core/services/notification.service';
import { BusService } from '../../core/services/bus.service';
import { DriverService } from '../../core/services/driver.service';
import { ConductorService } from '../../core/services/conductor.service';
import { LoungeService } from '../../core/services/lounge.service';
import { BusOwnerService } from '../../core/services/bus-owner.service';

interface NotificationDetails {
  id: string;
  type: 'bus' | 'driver' | 'conductor' | 'passenger' | 'lounge' | 'booking' | 'bus-owner' | 'lounge-owner';
  title: string;
  message: string;
  time: string;
  formData: any;
}

@Component({
  selector: 'app-notification-details',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent, NotificationPanelComponent, RouterModule],
  templateUrl: './notification-details.component.html',
  styleUrls: ['./notification-details.component.scss']
})
export class NotificationDetailsComponent implements OnInit {
  showNotificationPanel = false;
  showSuccessModal = false;
  successModalTitle = '';
  successModalMessage = '';
  showCancelModal = false;
  cancelModalTitle = '';
  cancelModalMessage = '';
  cancellationReasons: string[] = [];
  selectedReasons: { [key: string]: boolean } = {};
  otherReasonText = '';
  notification: NotificationDetails | null = null;
  busId: string = '';
  
  // Documents field for admin to fill
  adminDocuments: string = '';

  constructor(
    private route: ActivatedRoute, 
    private router: Router,
    public notificationService: NotificationService,
    private busService: BusService,
    private driverService: DriverService,
    private conductorService: ConductorService,
    private loungeService: LoungeService,
    private busOwnerService: BusOwnerService
  ) {}

  ngOnInit() {
    const id = this.route.snapshot.queryParamMap.get('id');
    const type = this.route.snapshot.queryParamMap.get('type') as 'bus' | 'driver' | 'conductor' | 'lounge' | 'bus-owner' | 'lounge-owner' | null;
    
    const navigation = this.router.getCurrentNavigation();
    const data = navigation?.extras?.state?.['data'] || history.state?.data;
    const stateType = navigation?.extras?.state?.['type'] || history.state?.type;
    
    if (id && (type || stateType)) {
      this.busId = id;
      const notifType = type || stateType;
      
      if (data) {
        // Use cached data
        this.mapDataToNotification(data, notifType);
      } else {
        // Fallback: fetch from backend
        this.loadDetails(id, notifType);
      }
    }
  }

  mapDataToNotification(data: any, type: string) {
    switch (type) {
      case 'bus':
        this.mapBusDataToNotification(data);
        break;
      case 'driver':
        this.mapDriverDataToNotification(data);
        break;
      case 'conductor':
        this.mapConductorDataToNotification(data);
        break;
      case 'lounge':
        this.mapLoungeDataToNotification(data);
        break;
      case 'bus-owner':
        this.mapBusOwnerDataToNotification(data);
        break;
      case 'lounge-owner':
        this.mapLoungeOwnerDataToNotification(data);
        break;
    }
  }

  mapBusDataToNotification(bus: any) {
    this.notification = {
      id: bus.id,
      type: 'bus',
      title: 'New Bus Added Request',
      message: `A new bus registration request has been submitted: Bus No: ${bus.permit_number || bus.bus_number}, Route: ${bus.custom_route_name || 'Not specified'}. Awaiting approval.`,
      time: this.getTimeAgo(new Date()),
      formData: {
        'Company': bus.company_name || 'N/A',
        'Phone Number': bus.business_phone || 'N/A',
        'NIC Number': bus.identify_or_incorporation_no || 'N/A',
        'Email': bus.business_email || 'N/A',
        'Permit Number': bus.permit_number || 'N/A',
        'Registered Number': bus.license_plate || 'N/A',
        'License plate from the permit': bus.license_plate || 'N/A',
        'Route via (Optional)': bus.custom_route_name || 'N/A',
        'Approved Fare': bus.fare_per_seat || 0,
        'Bus Type': bus.bus_type || 'Normal',
        'Validity period': '1/11/2025- 1/11/2026',
        'Seat numbers': bus.total_seats || 0
      }
    };
  }

  mapDriverDataToNotification(driver: any) {
    this.notification = {
      id: driver.id,
      type: 'driver',
      title: 'New Driver Added Request',
      message: `A new driver registration request has been submitted: ${driver.name}`,
      time: this.getTimeAgo(new Date()),
      formData: {
        'Name': driver.name || 'N/A',
        'Contact Number': driver.contact_number || 'N/A',
        'License Number': driver.license_number || 'N/A',
        'License Expiry Date': driver.license_expiry_date || 'N/A',
        'Experience Years': driver.experience_years || 0,
        'Employment Status': driver.status || 'N/A',
        'Hire Date': driver.hire_date || 'N/A',
        'Verification Notes': driver.verification_notes || 'None'
      }
    };
  }

  mapConductorDataToNotification(conductor: any) {
    this.notification = {
      id: conductor.id,
      type: 'conductor',
      title: 'New Conductor Added Request',
      message: `A new conductor registration request has been submitted: ${conductor.name}`,
      time: this.getTimeAgo(new Date()),
      formData: {
        'Name': conductor.name || 'N/A',
        'Contact Number': conductor.contact_number || 'N/A',
        'License Number': conductor.license_number || 'N/A',
        'License Expiry Date': conductor.license_expiry_date || 'N/A',
        'Experience Years': conductor.experience_years || 0,
        'Employment Status': conductor.status || 'N/A',
        'Hire Date': conductor.hire_date || 'N/A',
        'Verification Notes': conductor.verification_notes || 'None'
      }
    };
  }

  mapLoungeDataToNotification(lounge: any) {
    this.notification = {
      id: lounge.lounge_id,
      type: 'lounge',
      title: 'New Lounge Added Request',
      message: `A new lounge registration request has been submitted: ${lounge.lounge_name}`,
      time: this.getTimeAgo(new Date()),
      formData: {
        'Lounge Name': lounge.lounge_name || 'N/A',
        'Owner Name': lounge.lounge_owner || 'N/A',
        'Owner NIC': lounge.owner_nic || 'N/A',
        'Owner Email': lounge.owner_email || 'N/A',
        'Owner Contact': lounge.owner_contact || 'N/A',
        'Lounge Contact': lounge.lounge_contact || 'N/A',
        'Address': lounge.address || 'N/A',
        'Capacity': lounge.capacity || 0,
        'Price Per Hour': lounge.price_per_hour || 0,
        'Marketplace': lounge.marketplace || 'N/A',
        'Operational': lounge.operational ? 'Yes' : 'No'
      }
    };
  }

  mapBusOwnerDataToNotification(busOwner: any) {
    this.notification = {
      id: busOwner.id,
      type: 'bus-owner',
      title: 'New Bus Owner Added Request',
      message: `A new bus owner registration request has been submitted: ${busOwner.company_name || 'Unknown'}`,
      time: this.getTimeAgo(new Date()),
      formData: {
        'Company Name': busOwner.company_name || 'N/A',
        'Business Email': busOwner.business_email || 'N/A',
        'Business Phone': busOwner.business_phone || 'N/A',
        'NIC/Incorporation Number': busOwner.identity_or_incorporation_no || 'N/A',
        'Verification Status': busOwner.verification_status || 'pending'
      }
    };
  }

  mapLoungeOwnerDataToNotification(loungeOwner: any) {
    this.notification = {
      id: loungeOwner.id,
      type: 'lounge-owner',
      title: 'New Lounge Owner Added Request',
      message: `A new lounge owner registration request has been submitted: ${loungeOwner.manager_full_name || 'Unknown'}`,
      time: this.getTimeAgo(new Date()),
      formData: {
        'Manager Name': loungeOwner.manager_full_name || 'N/A',
        'Email': loungeOwner.email || 'N/A',
        'Contact Number': loungeOwner.contact_number || 'N/A',
        'NIC': loungeOwner.nic || 'N/A',
        'Business Name': loungeOwner.business_name || 'N/A',
        'Business License': loungeOwner.business_license || 'N/A',
        'Verification Status': loungeOwner.verification_status || 'pending'
      }
    };
  }

  loadDetails(id: string, type: string) {
    switch (type) {
      case 'bus':
        this.notificationService.getBusById(id).subscribe({
          next: (bus) => this.mapBusDataToNotification(bus),
          error: (err) => this.handleError(err)
        });
        break;
      case 'driver':
        this.notificationService.getDriverById(id).subscribe({
          next: (driver) => this.mapDriverDataToNotification(driver),
          error: (err) => this.handleError(err)
        });
        break;
      case 'conductor':
        this.notificationService.getConductorById(id).subscribe({
          next: (conductor) => this.mapConductorDataToNotification(conductor),
          error: (err) => this.handleError(err)
        });
        break;
      case 'lounge':
        this.notificationService.getLoungeById(id).subscribe({
          next: (lounge) => this.mapLoungeDataToNotification(lounge),
          error: (err) => this.handleError(err)
        });
        break;
      case 'bus-owner':
        this.notificationService.getBusOwnerById(id).subscribe({
          next: (busOwner) => this.mapBusOwnerDataToNotification(busOwner),
          error: (err) => this.handleError(err)
        });
        break;
      case 'lounge-owner':
        this.notificationService.getLoungeOwnerById(id).subscribe({
          next: (loungeOwner) => this.mapLoungeOwnerDataToNotification(loungeOwner),
          error: (err) => this.handleError(err)
        });
        break;
    }
  }

  handleError(err: any) {
    console.error('Error loading details:', err);
    this.router.navigate(['/dashboard']);
  }

  loadBusDetails(busId: string) {
    this.notificationService.getBusById(busId).subscribe({
      next: (bus) => {
        this.mapBusDataToNotification(bus);
      },
      error: (err) => {
        console.error('Error loading bus details:', err);
        this.router.navigate(['/dashboard']);
      }
    });
  }

  private getTimeAgo(date: Date): string {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffMins > 0) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
    return 'Just now';
  }

  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

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

  goBack() {
    this.router.navigate(['/dashboard']);
  }

  approveRequest(): void {
    if (this.notification && this.busId) {
      const type = this.notification.type;
      let approveObservable;
      
      // Prepare approval data with documents
      const approvalData = {
        documents: this.adminDocuments.trim()
      };
      
      switch (type) {
        case 'bus':
          approveObservable = this.notificationService.approveBus(this.busId, approvalData);
          this.successModalTitle = 'New Bus added Successfully!!!';
          this.successModalMessage = `The request to add a new bus has been approved.\nThe bus is now active in the system\nand ready for scheduling.\nThank you.`;
          break;
        case 'driver':
          approveObservable = this.notificationService.approveDriver(this.busId, approvalData);
          this.successModalTitle = 'New Driver added Successfully!!!';
          this.successModalMessage = `The request to add a new driver has been approved.\nThe driver is now active in the system\nand ready for assignment.\nThank you.`;
          break;
        case 'conductor':
          approveObservable = this.notificationService.approveConductor(this.busId, approvalData);
          this.successModalTitle = 'New Conductor added Successfully!!!';
          this.successModalMessage = `The request to add a new conductor has been approved.\nThe conductor is now active in the system\nand ready for assignment.\nThank you.`;
          break;
        case 'lounge':
          approveObservable = this.notificationService.approveLounge(this.busId, approvalData);
          this.successModalTitle = 'New Lounge added Successfully!!!';
          this.successModalMessage = `The request to add a new lounge has been approved.\nThe lounge is now active in the system\nand ready for booking.\nThank you.`;
          break;
        case 'bus-owner':
          approveObservable = this.notificationService.approveBusOwner(this.busId, { verification_documents: this.adminDocuments.trim() });
          this.successModalTitle = 'New Bus Owner added Successfully!!!';
          this.successModalMessage = `The request to add a new bus owner has been approved.\nThe bus owner is now verified in the system\nand ready to manage buses.\nThank you.`;
          break;
        case 'lounge-owner':
          approveObservable = this.notificationService.approveLoungeOwner(this.busId, { verification_notes: this.adminDocuments.trim() });
          this.successModalTitle = 'New Lounge Owner added Successfully!!!';
          this.successModalMessage = `The request to add a new lounge owner has been approved.\nThe lounge owner is now verified in the system\nand ready to manage lounges.\nThank you.`;
          break;
        default:
          return;
      }
      
      approveObservable.subscribe({
        next: () => {
          // Reload the respective service data after approval
          switch (type) {
            case 'bus':
              this.busService.loadBuses();
              break;
            case 'driver':
              this.driverService.loadDrivers();
              break;
            case 'conductor':
              this.conductorService.loadConductors();
              break;
            case 'lounge':
              this.loungeService.loadLounges();
              break;
            case 'bus-owner':
              this.busOwnerService.loadBusOwners();
              break;
            case 'lounge-owner':
              this.loungeService.loadLounges();
              break;
          }
          this.showSuccessModal = true;
        },
        error: (err) => {
          console.error(`Error approving ${type}:`, err);
          alert(`Failed to approve ${type}: ` + (err.error?.error || err.message));
        }
      });
    }
  }

  closeSuccessModal(): void {
    this.showSuccessModal = false;
  }

  sendApproval(): void {
    this.showSuccessModal = false;
    // Navigate to appropriate management page based on notification type
    if (this.notification && this.notification.type) {
      const type = this.notification.type;
      console.log('Navigating after approval, type:', type);
      switch (type) {
        case 'bus':
          this.router.navigate(['/bus-management']);
          break;
        case 'driver':
          this.router.navigate(['/driver-management']);
          break;
        case 'conductor':
          this.router.navigate(['/conductor-management']);
          break;
        case 'lounge':
          this.router.navigate(['/lounges-management']);
          break;
        case 'bus-owner':
          this.router.navigate(['/bus-owners']);
          break;
        case 'lounge-owner':
          this.router.navigate(['/lounges-management']);
          break;
        default:
          console.warn('Unknown notification type:', type);
          this.router.navigate(['/dashboard']);
      }
    } else {
      console.warn('No notification or notification type found, navigating to dashboard');
      this.router.navigate(['/dashboard']);
    }
  }

  cancelRequest(): void {
    if (this.notification) {
      const type = this.notification.type;
      switch (type) {
        case 'bus':
          this.cancelModalTitle = 'New Bus Addition Request Cancelled';
          this.cancelModalMessage = `The request to add a new bus has been cancelled.\nThe bus has not been added to the system.\nThank you.`;
          this.cancellationReasons = [
            'Bus details were missing or incorrect.',
            'A bus with the same identifier already exists.',
            'The request was not approved by the authorities.',
            'System could not process the bus addition.'
          ];
          break;
        case 'driver':
          this.cancelModalTitle = 'New Driver Addition Request Cancelled';
          this.cancelModalMessage = `The request to add a new driver has been cancelled.\nThe driver has not been added to the system.\nThank you.`;
          this.cancellationReasons = [
            'Driver details were missing or incorrect.',
            'A driver with the same license already exists.',
            'The request was not approved by the authorities.',
            'System could not process the driver addition.'
          ];
          break;
        case 'conductor':
          this.cancelModalTitle = 'New Conductor Addition Request Cancelled';
          this.cancelModalMessage = `The request to add a new conductor has been cancelled.\nThe conductor has not been added to the system.\nThank you.`;
          this.cancellationReasons = [
            'Conductor details were missing or incorrect.',
            'A conductor with the same license already exists.',
            'The request was not approved by the authorities.',
            'System could not process the conductor addition.'
          ];
          break;
        case 'lounge':
          this.cancelModalTitle = 'New Lounge Addition Request Cancelled';
          this.cancelModalMessage = `The request to add a new lounge has been cancelled.\nThe lounge has not been added to the system.\nThank you.`;
          this.cancellationReasons = [
            'Lounge details were missing or incorrect.',
            'A lounge with the same name already exists.',
            'The request was not approved by the authorities.',
            'System could not process the lounge addition.'
          ];
          break;
        case 'bus-owner':
          this.cancelModalTitle = 'New Bus Owner Addition Request Cancelled';
          this.cancelModalMessage = `The request to add a new bus owner has been cancelled.\nThe bus owner has not been added to the system.\nThank you.`;
          this.cancellationReasons = [
            'Business details were missing or incorrect.',
            'A bus owner with the same details already exists.',
            'The request was not approved by the authorities.',
            'System could not process the bus owner addition.'
          ];
          break;
        case 'lounge-owner':
          this.cancelModalTitle = 'New Lounge Owner Addition Request Cancelled';
          this.cancelModalMessage = `The request to add a new lounge owner has been cancelled.\nThe lounge owner has not been added to the system.\nThank you.`;
          this.cancellationReasons = [
            'Manager details were missing or incorrect.',
            'A lounge owner with the same details already exists.',
            'The request was not approved by the authorities.',
            'System could not process the lounge owner addition.'
          ];
          break;
        default:
          this.cancelModalTitle = 'Request Cancelled';
          this.cancelModalMessage = `The request has been cancelled.\nThank you.`;
          this.cancellationReasons = ['Request was not approved.'];
      }
    } else {
      this.cancelModalTitle = 'Request Cancelled';
      this.cancelModalMessage = `The request has been cancelled.\nThank you.`;
      this.cancellationReasons = ['Request was not approved.'];
    }
    
    // Reset selection
    this.selectedReasons = {};
    this.otherReasonText = '';
    
    this.showCancelModal = true;
  }

  closeCancelModal(): void {
    this.showCancelModal = false;
  }

  sendCancellation(): void {
    if (this.busId && this.notification) {
      const type = this.notification.type;
      let rejectObservable;
      
      switch (type) {
        case 'bus':
          rejectObservable = this.notificationService.rejectBus(this.busId);
          break;
        case 'driver':
          rejectObservable = this.notificationService.rejectDriver(this.busId);
          break;
        case 'conductor':
          rejectObservable = this.notificationService.rejectConductor(this.busId);
          break;
        case 'lounge':
          rejectObservable = this.notificationService.rejectLounge(this.busId);
          break;
        case 'bus-owner':
          rejectObservable = this.notificationService.rejectBusOwner(this.busId, { verification_documents: this.otherReasonText.trim() });
          break;
        case 'lounge-owner':
          rejectObservable = this.notificationService.rejectLoungeOwner(this.busId, { verification_notes: this.otherReasonText.trim() });
          break;
        default:
          return;
      }
      
      rejectObservable.subscribe({
        next: () => {
          this.showCancelModal = false;
          console.log(`${type} rejected`, {
            reasons: this.selectedReasons,
            other: this.otherReasonText
          });
          // Navigate based on type
          switch (type) {
            case 'bus':
              this.router.navigate(['/bus-management']);
              break;
            case 'driver':
              this.router.navigate(['/driver-management']);
              break;
            case 'conductor':
              this.router.navigate(['/conductor-management']);
              break;
            case 'lounge':
              this.router.navigate(['/lounges-management']);
              break;
            case 'bus-owner':
              this.router.navigate(['/bus-owners']);
              break;
            case 'lounge-owner':
              this.router.navigate(['/lounges-management']);
              break;
            default:
              this.router.navigate(['/dashboard']);
          }
        },
        error: (err) => {
          console.error(`Error rejecting ${type}:`, err);
          alert(`Failed to reject ${type}: ` + (err.error?.error || err.message));
        }
      });
    }
  }

  getFormFields(): { label: string; value: any; key: string }[] {
    if (!this.notification) return [];
    
    return Object.entries(this.notification.formData).map(([key, value]) => ({
      label: this.formatLabel(key),
      value: value,
      key: key
    }));
  }

  private formatLabel(key: string): string {
    return key
      .replace(/([A-Z])/g, ' $1')
      .replace(/^./, str => str.toUpperCase())
      .trim();
  }

  isArray(value: any): boolean {
    return Array.isArray(value);
  }
}
