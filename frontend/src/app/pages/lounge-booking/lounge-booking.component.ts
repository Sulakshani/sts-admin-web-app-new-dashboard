import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { LoungeBookingService } from '../../core/services/lounge-booking.service';
import { LoungeBooking } from '../../core/models/lounge-booking.model';
import { ChartData, ChartOptions } from 'chart.js';
import { BaseChartDirective } from 'ng2-charts';
import { Chart, registerables } from 'chart.js';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'app-lounge-booking',
  standalone: true,
  imports: [CommonModule, FormsModule, BaseChartDirective, NotificationPanelComponent, NavbarComponent, RouterModule],
  templateUrl: './lounge-booking.component.html',
  styleUrls: ['./lounge-booking.component.scss']
})
export class LoungeBookingComponent implements OnInit {
  currentPage = 'lounge-booking';
  isBrowser!: boolean;
  showNotificationPanel = false;
  showProfileMenu = false;


  bookings: LoungeBooking[] = [];
  filtered: LoungeBooking[] = [];
  searchTerm = '';

  paymentFilter: 'All' | 'pending' | 'paid' | 'failed' = 'All';
  statusFilter: 'All' | 'confirmed' | 'pending' | 'cancelled' | 'completed' = 'All';

  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';

  // Chart data
  payStatusCounts: Record<string, number> = {};
  bookStatusCounts: Record<string, number> = {};
  revenueMonths: number[] = [];
  months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];

  // Monthly (current month) paid revenue grouped by lounge
  private currentYear = new Date().getFullYear();
  private currentMonth = new Date().getMonth(); // 0-11
  revenueByLounge: { name: string; total: number }[] = [];
  private colors: string[] = ['#4caf50', '#2196f3', '#ff9800', '#e91e63', '#9c27b0', '#00bcd4', '#8bc34a'];

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
          stepSize: 1
        }
      }
    }
  };

  // Remove pie chart helpers
  // getRevenuePieBackground(): string {
    exportLoungeBookingHistoryPdf(): void {
      const doc = new jsPDF({ orientation: 'landscape' });
      doc.setFontSize(16);
      doc.text('Lounge Booking History', 14, 16);

      const tableHead = [[
        'Booking ID', 'Passenger Name', 'Phone', 'Ref Num', 'Lounge Name', 'Date', 'Time', 'Guests', 'Features', 'Total Amount', 'Payment', 'Status'
      ]];
      const tableBody = this.filtered.map(b => [
        b.lounge_booking_id,
        b.passenger_name || '',
        b.passenger_phone || '',
        b.booking_reference || '',
        b.lounge_name,
        new Date(b.scheduled_arrival).toLocaleDateString(),
        new Date(b.scheduled_arrival).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'}),
        b.number_of_guests?.toString() || '0',
        (b.selected_amenities || []).join(', '),
        `$${b.total_amount}`,
        b.payment_status,
        b.status
      ]);

      autoTable(doc, {
        head: tableHead,
        body: tableBody,
        startY: 22,
        styles: { fontSize: 8 },
        headStyles: { fillColor: [59, 130, 246] }
      });

      doc.save('lounge-booking-history.pdf');
    }
  //   if (this.revenueByLounge.length === 0) return 'conic-gradient(#ccc 0% 100%)';
  //   const total = this.revenueByLounge.reduce((sum, item) => sum + item.total, 0);
  //   let currentPercent = 0;
  //   const gradients = this.revenueByLounge.map((item, index) => {
  //     const percent = (item.total / total) * 100;
  //     const start = currentPercent;
  //     const end = currentPercent + percent;
  //     currentPercent = end;
  //     const color = this.colors[index % this.colors.length];
  //     return `${color} ${start}% ${end}%`;
  //   });
  //   return `conic-gradient(${gradients.join(', ')})`;
  // }
  //
  // getColorForLounge(name: string): string {
  //   const index = this.revenueByLounge.findIndex(item => item.name === name);
  //   return this.colors[index % this.colors.length];
  // }

  constructor(private router: Router, private svc: LoungeBookingService, @Inject(PLATFORM_ID) private platformId: Object, public notificationService: NotificationService) {
    this.isBrowser = isPlatformBrowser(this.platformId);
    if (this.isBrowser) {
      Chart.register(...registerables);
    }
  }

  ngOnInit(): void {
    this.svc.bookings$.subscribe(bs => {
      this.bookings = bs;
      this.applyFilters();
      this.refreshCharts();
      this.refreshRevenueByLounge();
      this.updateChartData();
    });
  }

  onNavigate(page: string) {
    this.currentPage = page;
    this.router.navigate([`/${page}`]);
  }
  onLogout() {
    localStorage.removeItem('token');
    this.router.navigate(['/']);
  }

  applyFilters() {
    const q = this.searchTerm.trim().toLowerCase();
    this.filtered = this.bookings.filter(b => {
      const matchesSearch = !q || [
        b.lounge_booking_id,
        b.bus_booking_id || '',
        b.passenger_name || '',
        b.passenger_phone || '',
        b.booking_reference || '',
        b.lounge_name,
        b.payment_status,
        b.status,
        b.scheduled_arrival,
        b.number_of_guests.toString(),
        b.pricing_type || '',
        (b.selected_amenities || []).join(' '),
        b.product_name || '',
        b.created_at || ''
      ].some(x => x && x.toLowerCase().includes(q)) ||
      b.total_amount.toString().includes(q);

      const matchesPay = this.paymentFilter === 'All' || b.payment_status === this.paymentFilter;
      const matchesStatus = this.statusFilter === 'All' || b.status === this.statusFilter;
      return matchesSearch && matchesPay && matchesStatus;
    });
    this.applySorting();
  }

  clearSearch() { this.searchTerm = ''; this.applyFilters(); }

  // Modal state
  showModal = false;
  isEditMode = false;
  formBooking: Partial<LoungeBooking> = {};
  
  availableFeatures = ['Premium meals', 'Express loundary', 'cargo storage', 'spa service', 'personal assist', 'Airport transfer', 'Tuk tuk'];

  openAddModal() {
    this.isEditMode = false;
    this.formBooking = {
      lounge_name: '',
      passenger_name: '',
      passenger_phone: '',
      booking_reference: 'LNG-' + Math.random().toString(36).substring(2, 8).toUpperCase(),
      scheduled_arrival: '',
      pricing_type: '1_hour',
      number_of_guests: 1,
      selected_amenities: [],
      product_name: '',
      booking_type: 'standalone',
      total_amount: 0,
      payment_status: 'pending',
      status: 'pending'
    };
    this.showModal = true;
  }

  openEditModal(b: LoungeBooking) {
    this.isEditMode = true;
    this.formBooking = { ...b };
    
    // Format datetime for input (YYYY-MM-DDTHH:mm)
    if (this.formBooking.scheduled_arrival) {
      try {
        this.formBooking.scheduled_arrival = new Date(this.formBooking.scheduled_arrival).toISOString().slice(0, 16);
      } catch (e) {
        console.error('Invalid date format', e);
      }
    }

    // Ensure selected_amenities is an array
    if (!this.formBooking.selected_amenities) {
      this.formBooking.selected_amenities = [];
    }
    this.showModal = true;
  }

  closeModal() {
    this.showModal = false;
  }

  // Validate lounge booking form - all required fields must be filled except booking_reference and product_name
  isLoungeBookingFormValid(): boolean {
    return !!(
      this.formBooking.passenger_name?.trim() &&
      this.formBooking.passenger_phone?.trim() &&
      this.formBooking.lounge_name?.trim() &&
      this.formBooking.scheduled_arrival &&
      this.formBooking.pricing_type?.trim() &&
      this.formBooking.number_of_guests !== null &&
      this.formBooking.number_of_guests !== undefined &&
      this.formBooking.number_of_guests > 0 &&
      this.formBooking.total_amount !== null &&
      this.formBooking.total_amount !== undefined &&
      this.formBooking.total_amount >= 0
    );
  }

  saveBooking() {
    // Validate all required fields
    if (!this.isLoungeBookingFormValid()) {
      alert('Please fill all required fields before saving.');
      return;
    }

    if (this.isEditMode) {
      // Convert datetime-local format back to ISO string
      const bookingToUpdate = { 
        ...this.formBooking,
        scheduled_arrival: this.formBooking.scheduled_arrival 
          ? new Date(this.formBooking.scheduled_arrival).toISOString()
          : new Date().toISOString()
      } as LoungeBooking;
      
      this.svc.update(bookingToUpdate).subscribe({
        next: () => {
          console.log('✓ Lounge booking updated successfully');
          this.closeModal();
          alert('Lounge booking updated successfully!');
        },
        error: (err) => {
          console.error('✗ Error updating booking:', err);
          alert('Error updating booking: ' + (err.error?.error || err.message || 'Unknown error'));
        }
      });
    } else {
      const newBooking = { 
        ...this.formBooking,
        bus_booking_id: null,
        scheduled_arrival: this.formBooking.scheduled_arrival 
          ? new Date(this.formBooking.scheduled_arrival).toISOString()
          : new Date().toISOString(),
        created_at: new Date().toISOString()
      } as LoungeBooking;
      
      this.svc.add(newBooking).subscribe({
        next: () => {
          console.log('✓ Lounge booking added successfully');
          this.closeModal();
          alert('Lounge booking added successfully!');
        },
        error: (err) => {
          console.error('✗ Error creating booking:', err);
          alert('Error creating booking: ' + (err.error?.error || err.message || 'Unknown error'));
        }
      });
    }
  }

  toggleFeature(feature: string) {
    const features = this.formBooking.selected_amenities || [];
    if (features.includes(feature)) {
      this.formBooking.selected_amenities = features.filter(f => f !== feature);
    } else {
      this.formBooking.selected_amenities = [...features, feature];
    }
  }

  isFeatureSelected(feature: string): boolean {
    return (this.formBooking.selected_amenities || []).includes(feature);
  }

  deleteBooking(b: LoungeBooking) {
    const ok = confirm(`Delete booking ${b.lounge_booking_id}?`);
    if (ok) {
      this.svc.delete(b.lounge_booking_id).subscribe({
        next: () => this.svc.loadBookings(),
        error: (err) => console.error('Error deleting booking:', err)
      });
    }
  }

  changePaymentStatus(b: LoungeBooking, v: 'pending'|'paid'|'failed') {
    this.svc.updatePaymentStatus(b.lounge_booking_id, v).subscribe({
      next: () => this.svc.loadBookings(),
      error: (err) => console.error('Error updating payment status:', err)
    });
  }
  
  changeBookingStatus(b: LoungeBooking, v: 'confirmed'|'pending'|'cancelled'|'completed') {
    this.svc.updateBookingStatus(b.lounge_booking_id, v).subscribe({
      next: () => this.svc.loadBookings(),
      error: (err) => console.error('Error updating booking status:', err)
    });
  }

  refreshCharts() {
    this.payStatusCounts = this.svc.countByPaymentStatus();
    this.bookStatusCounts = this.svc.countByBookingStatus();
    this.revenueMonths = this.svc.monthlyRevenue(new Date().getFullYear());
  }
  //notification panel
  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

  // Aggregate total revenue by lounge
  refreshRevenueByLounge() {
    const acc = new Map<string, number>();
    this.bookings.forEach(b => {
      acc.set(b.lounge_name, (acc.get(b.lounge_name) || 0) + b.total_amount);
    });
    this.revenueByLounge = Array.from(acc.entries())
      .map(([name, total]) => ({ name, total }))
      .sort((a, b) => b.total - a.total);
    this.updateChartData();
  }

  updateChartData() {
    this.barChartData = [
      {
        labels: ['Paid', 'Pending', 'Failed'],
        datasets: [{
          data: [this.payStatusCounts['paid'] || 0, this.payStatusCounts['pending'] || 0, this.payStatusCounts['failed'] || 0],
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533'],
          borderWidth: 0.25
        }]
      },
      {
        labels: ['Confirmed', 'Pending', 'Cancelled', 'Completed'],
        datasets: [{
          data: [this.bookStatusCounts['confirmed'] || 0, this.bookStatusCounts['pending'] || 0, this.bookStatusCounts['cancelled'] || 0, this.bookStatusCounts['completed'] || 0],
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#6db9f8ff'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#6db9f8ff'],
          borderWidth: 0.25
        }]
      },
      {
        labels: this.revenueByLounge.map(item => item.name),
        datasets: [{
          data: this.revenueByLounge.map(item => item.total),
          backgroundColor: ['#0046FF', '#a3a3a3', '#FAA533', '#6db9f8ff'],
          borderColor: ['#0046FF', '#a3a3a3', '#FAA533', '#6db9f8ff'],
          borderWidth: 0.25
        }]
      }
    ];
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

  // Bar chart helpers
  getRevenueMaxForBars(): number {
    return Math.max(1, ...this.revenueByLounge.map(x => x.total));
  }
  getRevenueBarHeight(total: number): number {
    const max = this.getRevenueMaxForBars();
    return (total / max) * 100;
  }

  // Simple inline charts helpers
  getPayCount(key: 'Paid'|'Pending'|'Failed') { return this.payStatusCounts[key] || 0; }
  getBookCount(key: 'Confirmed'|'Pending'|'Cancelled'|'Completed') { return this.bookStatusCounts[key] || 0; }
  getRevenueMax() { return Math.max(1, ...this.revenueMonths); }

  // Helpers to avoid Math.* in template
  private max3(a: number, b: number, c: number): number { return Math.max(1, a, b, c); }
  private max4(a: number, b: number, c: number, d: number): number { return Math.max(1, a, b, c, d); }
  getPayBarHeight(key: 'Paid'|'Pending'|'Failed'): number {
    const paid = this.getPayCount('Paid');
    const pending = this.getPayCount('Pending');
    const failed = this.getPayCount('Failed');
    const max = this.max3(paid, pending, failed) || 1;
    const val = this.getPayCount(key);
    return (val / max) * 100;
  }
  getBookBarHeight(key: 'Confirmed'|'Pending'|'Cancelled'|'Completed'): number {
    const c = this.getBookCount('Confirmed');
    const p = this.getBookCount('Pending');
    const x = this.getBookCount('Cancelled');
    const d = this.getBookCount('Completed');
    const max = this.max4(c, p, x, d) || 1;
    const val = this.getBookCount(key);
    return (val / max) * 100;
  }

  // Build SVG polyline points for revenue chart
  getRevenuePoints(): string {
    const width = 600, height = 220, pad = 30;
    const max = Math.max(1, ...this.revenueMonths);
    const stepX = (width - 2 * pad) / (this.months.length - 1);
    return this.revenueMonths
      .map((v, i) => {
        const x = pad + i * stepX;
        const y = height - pad - (v / max) * (height - 2 * pad);
        return `${x},${y}`;
      })
      .join(' ');
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
      let aValue: any;
      let bValue: any;

      switch (this.sortColumn) {
        case 'lounge_booking_id':
          aValue = a.lounge_booking_id?.toLowerCase() || '';
          bValue = b.lounge_booking_id?.toLowerCase() || '';
          break;
        case 'passenger_name':
          aValue = a.passenger_name?.toLowerCase() || '';
          bValue = b.passenger_name?.toLowerCase() || '';
          break;
        case 'passenger_phone':
          aValue = a.passenger_phone?.toLowerCase() || '';
          bValue = b.passenger_phone?.toLowerCase() || '';
          break;
        case 'booking_reference':
          aValue = a.booking_reference?.toLowerCase() || '';
          bValue = b.booking_reference?.toLowerCase() || '';
          break;
        case 'lounge_name':
          aValue = a.lounge_name.toLowerCase();
          bValue = b.lounge_name.toLowerCase();
          break;
        case 'scheduled_arrival':
          aValue = new Date(a.scheduled_arrival).getTime();
          bValue = new Date(b.scheduled_arrival).getTime();
          break;
        case 'pricing_type':
          aValue = a.pricing_type?.toLowerCase() || '';
          bValue = b.pricing_type?.toLowerCase() || '';
          break;
        case 'number_of_guests':
          aValue = a.number_of_guests;
          bValue = b.number_of_guests;
          break;
        case 'selected_amenities':
          aValue = (a.selected_amenities || []).join(', ').toLowerCase();
          bValue = (b.selected_amenities || []).join(', ').toLowerCase();
          break;
        case 'product_name':
          aValue = a.product_name?.toLowerCase() || '';
          bValue = b.product_name?.toLowerCase() || '';
          break;
        case 'total_amount':
          aValue = a.total_amount;
          bValue = b.total_amount;
          break;
        case 'payment_status':
          aValue = a.payment_status.toLowerCase();
          bValue = b.payment_status.toLowerCase();
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

  // Pie chart helpers for revenue by lounge
  getRevenuePieBackground(): string {
    if (this.revenueByLounge.length === 0) return 'conic-gradient(#ccc 0% 100%)';
    const total = this.revenueByLounge.reduce((sum, item) => sum + item.total, 0);
    let currentPercent = 0;
    const gradients = this.revenueByLounge.map((item, index) => {
      const percent = (item.total / total) * 100;
      const start = currentPercent;
      const end = currentPercent + percent;
      currentPercent = end;
      const color = this.colors[index % this.colors.length];
      return `${color} ${start}% ${end}%`;
    });
    return `conic-gradient(${gradients.join(', ')})`;
  }

  getColorForLounge(name: string): string {
    const index = this.revenueByLounge.findIndex(item => item.name === name);
    return this.colors[index % this.colors.length];
  }
}
