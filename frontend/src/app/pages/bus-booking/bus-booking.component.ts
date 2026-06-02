import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { BusBookingService } from '../../core/services/bus-booking.service';
import { BusBooking } from '../../core/models/bus-booking.model';
import { ChartData, ChartOptions } from 'chart.js';
import { BaseChartDirective } from 'ng2-charts';
import { Chart, registerables } from 'chart.js';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-bus-booking',
  standalone: true,
  imports: [CommonModule, FormsModule, BaseChartDirective, NotificationPanelComponent, NavbarComponent, RouterModule],
  templateUrl: './bus-booking.component.html',
  styleUrls: ['./bus-booking.component.scss']
})
export class BusBookingComponent implements OnInit {
  exportBusBookingHistoryPdf(): void {
    const doc = new jsPDF({ orientation: 'landscape' });
    doc.setFontSize(16);
    doc.text('Bus Booking History', 14, 16);

    const tableHead = [[
      'Booking ID', 'BusID', 'Passenger Name', 'Passenger Phone', 'Ref NUM', 'Route', 'Date & Time', 'Bus Type', 'Seat No', 'Total Fare', 'Payment Status', 'Booking Status'
    ]];
    const tableBody = this.filtered.map(b => [
      b.booking_id,
      b.bus_number || '-',
      b.passenger_name,
      b.passenger_phone || '-',
      b.booking_reference || '-',
      b.route,
      new Date(b.departure_datetime).toLocaleString(),
      b.bus_type || '-',
      b.seat_number || '-',
      `$${b.total_fare}`,
      b.payment_status,
      b.booking_status
    ]);

    autoTable(doc, {
      head: tableHead,
      body: tableBody,
      startY: 22,
      styles: { fontSize: 8 },
      headStyles: { fillColor: [59, 130, 246] }
    });

    doc.save('bus-booking-history.pdf');
  }
  
  currentPage = 'bus-booking';
  isBrowser!: boolean;
  showNotificationPanel = false;
  showProfileMenu = false;

  bookings: BusBooking[] = [];
  filtered: BusBooking[] = [];
  searchTerm = '';

  paymentFilter: 'All' | 'Pending' | 'Paid' | 'Failed' | 'Refunded' = 'All';
  statusFilter: 'All' | 'Confirmed' | 'Pending' | 'Cancelled' | 'Completed' = 'All';

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // Chart data
  payStatusCounts: Record<string, number> = {};
  bookStatusCounts: Record<string, number> = {};
  revenueMonths: number[] = [];
  months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];

  // Chart.js data
  barChartData: any[] = [];
  barChartOptions: any = {
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
          maxTicksLimit: 8
        }
      }
    }
  };

  // Modal properties
  showEditModal = false;
  isEditMode = false;
  selectedBooking: BusBooking | null = null;
  formSeatNumbers = '';

  // Expose Math to template
  Math = Math;

  constructor(private router: Router, private svc: BusBookingService, @Inject(PLATFORM_ID) private platformId: Object, public notificationService: NotificationService) {
    this.isBrowser = isPlatformBrowser(this.platformId);
    if (this.isBrowser) {
      Chart.register(...registerables);
    }
  }

  ngOnInit(): void {
    // Subscribe to bookings observable for reactive updates
    this.svc.bookings$.subscribe(bs => {
      console.log('Bookings updated in component:', bs.length);
      this.bookings = bs;
      this.applyFilters();
      this.refreshCharts();
      this.updateChartData();
    });
  }

  goDashboard() {
    this.currentPage = 'dashboard';
    this.router.navigate(['/dashboard']);
  }

  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

  applyFilters() {
    const q = this.searchTerm.trim().toLowerCase();
    this.filtered = this.bookings.filter(b => {
      const matchesSearch = !q || [
        b.booking_id,
        b.scheduled_trip_id || '',
        b.bus_id || '',
        b.passenger_name,
        b.passenger_phone || '',
        b.booking_reference || '',
        b.bus_number,
        b.license_plate || '',
        b.bus_type || '',
        b.route,
        b.seat_number || '',
        b.departure_datetime || '',
        b.payment_status || '',
        b.booking_status || '',
        b.created_at || ''
      ].some(x => x.toLowerCase().includes(q)) ||
      b.total_fare.toString().includes(q) || b.number_of_seats.toString().includes(q);

      const matchesPay = this.paymentFilter === 'All' || b.payment_status?.toLowerCase() === this.paymentFilter.toLowerCase();
      const matchesStatus = this.statusFilter === 'All' || b.booking_status?.toLowerCase() === this.statusFilter.toLowerCase();
      return matchesSearch && matchesPay && matchesStatus;
    });
    
    // Re-apply sorting after filtering
    if (this.sortColumn) {
      this.applySorting();
    }
  }

  clearSearch() { this.searchTerm = ''; this.applyFilters(); }

  viewBooking(b: BusBooking) { alert(`View ${b.booking_id}`); }
  
  openAddModal() {
    this.isEditMode = false;
    this.selectedBooking = {
      booking_id: `BBK-${Math.floor(Math.random() * 10000)}`,
      scheduled_trip_id: '',
      bus_id: '',
      passenger_name: '',
      passenger_phone: '',
      booking_reference: '',
      route: '',
      departure_datetime: '',
      bus_type: '',
      seat_number: '',
      total_fare: 0,
      payment_status: 'pending',
      booking_status: 'pending',
      created_at: new Date().toISOString(),
      bus_number: '',
      license_plate: '',
      number_of_seats: 0
    };
    this.formSeatNumbers = '';
    this.showEditModal = true;
  }

  updateBooking(b: BusBooking) {
    this.isEditMode = true;
    this.selectedBooking = { ...b };
    this.formSeatNumbers = b.seat_number || '';
    this.showEditModal = true;
  }
  deleteBooking(b: BusBooking) {
    const ok = confirm(`Delete booking ${b.booking_id}?`);
    if (ok) this.svc.delete(b.booking_id);
  }

  refreshCharts() {
    this.payStatusCounts = this.svc.countByPaymentStatus();
    this.bookStatusCounts = this.svc.countByBookingStatus();
    this.revenueMonths = this.svc.monthlyRevenue(new Date().getFullYear());
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

  // Helpers for simple CSS charts
  getPayCount(key: 'Paid'|'Pending'|'Failed'|'Refunded') { return this.payStatusCounts[key] || 0; }
  getBookCount(key: 'Confirmed'|'Pending'|'Cancelled'|'Completed') { return this.bookStatusCounts[key] || 0; }
  getRevenueMax() { return Math.max(1, ...this.revenueMonths); }

  getRevenuePoints(): string {
    const max = Math.max(1, ...this.revenueMonths);
    return this.revenueMonths.map((v, i) => `${i * 50 + 50},${260 - (v / max) * 200}`).join(' ');
  }

  getYAxisLabels(): { value: string, y: number }[] {
    const max = Math.max(1, ...this.revenueMonths);
    const steps = [0, 0.25, 0.5, 0.75, 1];
    return steps.map(f => {
      const val = f * max;
      const y = 260 - (val / max) * 200;
      return { value: Math.round(val).toString(), y };
    });
  }

  onPaymentStatusChange(booking: BusBooking): void {
    console.log('Payment status changed:', booking);
    // Update charts after payment status change
    this.refreshCharts();
    this.updateChartData();
    // Add your update logic here, e.g., call API to update backend
  }

  onBookingStatusChange(booking: BusBooking): void {
    console.log('Booking status changed:', booking);
    // Update charts after booking status change
    this.refreshCharts();
    this.updateChartData();
    // Add your update logic here, e.g., call API to update backend
  }

  updateChartData() {
    this.barChartData = [
      {
        labels: ['Paid', 'Pending', 'Failed', 'Refunded', 'Collect on Bus'],
        datasets: [{
          data: [
            this.payStatusCounts['paid'] || 0,
            this.payStatusCounts['pending'] || 0,
            this.payStatusCounts['failed'] || 0,
            this.payStatusCounts['refunded'] || 0,
            this.payStatusCounts['collect_on_bus'] || 0
          ],
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B', '#10B981'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B', '#10B981'],
          borderWidth: 0.25
        }]
      },
      {
        labels: ['Confirmed', 'Pending', 'Cancelled', 'Completed'],
        datasets: [{
          data: [
            this.bookStatusCounts['confirmed'] || 0,
            this.bookStatusCounts['pending'] || 0,
            this.bookStatusCounts['cancelled'] || 0,
            this.bookStatusCounts['completed'] || 0
          ],
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
          borderWidth: 0.25
        }]
      },
      {
        labels: this.months,
        datasets: [{
          data: this.revenueMonths,
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B', '#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B', '#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B', '#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B', '#0046FF', '#a3a3a3', '#FAA533', '#FF6B6B'],
          borderWidth: 0.25
        }]
      }
    ];
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
    this.filtered = [...this.filtered].sort((a, b) => {
      let aValue: any = '';
      let bValue: any = '';

      // Helper to safely get string value
      const getStr = (val: any) => (val || '').toString().toLowerCase();

      switch (this.sortColumn) {
        case 'booking_id':
          aValue = getStr(a.booking_id);
          bValue = getStr(b.booking_id);
          break;
        case 'scheduled_trip_id':
          aValue = getStr(a.scheduled_trip_id);
          bValue = getStr(b.scheduled_trip_id);
          break;
        case 'bus_number':
          aValue = getStr(a.bus_number);
          bValue = getStr(b.bus_number);
          break;
        case 'passenger_name':
          aValue = getStr(a.passenger_name);
          bValue = getStr(b.passenger_name);
          break;
        case 'passenger_phone':
          aValue = getStr(a.passenger_phone);
          bValue = getStr(b.passenger_phone);
          break;
        case 'booking_reference':
          aValue = getStr(a.booking_reference);
          bValue = getStr(b.booking_reference);
          break;
        case 'route':
          aValue = getStr(a.route);
          bValue = getStr(b.route);
          break;
        case 'departure_datetime':
          aValue = new Date(a.departure_datetime).getTime();
          bValue = new Date(b.departure_datetime).getTime();
          break;
        case 'bus_type':
          aValue = getStr(a.bus_type);
          bValue = getStr(b.bus_type);
          break;
        case 'seat_number':
          aValue = getStr(a.seat_number);
          bValue = getStr(b.seat_number);
          break;
        case 'total_fare':
          aValue = a.total_fare;
          bValue = b.total_fare;
          break;
        case 'payment_status':
          aValue = getStr(a.payment_status);
          bValue = getStr(b.payment_status);
          break;
        case 'booking_status':
          aValue = getStr(a.booking_status);
          bValue = getStr(b.booking_status);
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

  // Modal methods
  closeEditModal(): void {
    this.showEditModal = false;
    this.selectedBooking = null;
  }

  // Validate bus booking form - all required fields must be filled except booking_reference
  isBusBookingFormValid(): boolean {
    if (!this.selectedBooking) return false;
    
    return !!(
      this.selectedBooking.passenger_name?.trim() &&
      this.selectedBooking.route?.trim() &&
      this.selectedBooking.departure_datetime &&
      this.formSeatNumbers?.trim() &&
      this.selectedBooking.total_fare !== null &&
      this.selectedBooking.total_fare !== undefined &&
      this.selectedBooking.total_fare >= 0 &&
      this.selectedBooking.bus_type
    );
  }

  saveBooking(): void {
    console.log('saveBooking called, selectedBooking:', this.selectedBooking);
    if (!this.selectedBooking) {
      console.log('No selected booking, returning');
      return;
    }
    
    // Validate all required fields
    if (!this.isBusBookingFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }
    
    this.selectedBooking.seat_number = this.formSeatNumbers;
    const seatCount = this.formSeatNumbers.split(',').filter(s => s.trim()).length;
    this.selectedBooking.number_of_seats = seatCount;
    
    console.log('Saving booking, isEditMode:', this.isEditMode);
    console.log('Booking data:', this.selectedBooking);
    
    if (this.isEditMode) {
      console.log('Calling update service...');
      this.svc.update(this.selectedBooking).subscribe({
        next: (result) => {
          console.log('✓ Update complete, data reloaded:', result);
          this.closeEditModal();
          alert('Bus booking updated successfully!');
        },
        error: (err) => {
          console.error('✗ Error updating booking:', err);
          alert('Error updating booking: ' + (err.error?.error || err.message));
        }
      });
    } else {
      console.log('Calling add service...');
      this.svc.add(this.selectedBooking).subscribe({
        next: (result) => {
          console.log('✓ Add complete, data reloaded:', result);
          this.closeEditModal();
          alert('Bus booking added successfully!');
        },
        error: (err) => {
          console.error('✗ Error saving booking:', err);
          alert('Error creating booking: ' + (err.error?.error || err.message));
        }
      });
    }
  }
}
