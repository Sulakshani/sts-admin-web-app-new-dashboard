import { Component, OnInit, Inject, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterModule } from '@angular/router';
import { NavbarComponent } from '../../shared/components/navbar/navbar.component';
import { NotificationPanelComponent } from '../../shared/components/notification-panel/notification-panel.component';
import { PassengerService } from '../../core/services/passenger.service';
import { BusBookingService } from '../../core/services/bus-booking.service';
import { LoungeBookingService } from '../../core/services/lounge-booking.service';
import { Passenger } from '../../core/models/passenger.model';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import { NotificationService } from '../../core/services/notification.service';
import { BaseChartDirective } from 'ng2-charts';
import { Chart, ChartData, ChartOptions, registerables } from 'chart.js';

@Component({
  selector: 'app-passenger-management',
  standalone: true,
  imports: [CommonModule, FormsModule, NavbarComponent, NotificationPanelComponent, RouterModule, BaseChartDirective],
  templateUrl: './passenger-management.component.html',
  styleUrls: ['./passenger-management.component.scss']
})
export class PassengerManagementComponent implements OnInit {
  isBrowser!: boolean;
  passengers: Passenger[] = [];
  filteredPassengers: Passenger[] = [];
  searchTerm = '';
  currentPage: string = 'passenger-management';  // default page
  showNotificationPanel = false;
  showProfileMenu = false;
  // Sorting properties
  sortColumn: string = '';
  sortDirection: 'asc' | 'desc' = 'asc';


  monthlyCounts: number[] = [];
  months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
  routeBookingCounts: { route: string; count: number; percentage: number }[] = [];
  
  // Chart.js data
  routeChartData: ChartData<'bar'> = {
    labels: [],
    datasets: [{
      data: [],
      backgroundColor: ['#0046FF', '#9CA3AF', '#FB923C', '#60A5FA', '#0046FF', '#6B7280', '#FBBF24'],
      borderColor: ['#0046FF', '#9CA3AF', '#FB923C', '#60A5FA', '#0046FF', '#6B7280', '#FBBF24'],
      borderWidth: 1
    }]
  };
  
  routeChartOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false
      }
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: {
          color: '#E5E7EB'
        },
        ticks: {
          stepSize: 1300,
          callback: function(value) {
            return value.toLocaleString();
          }
        }
      },
      x: {
        grid: {
          display: false
        }
      }
    }
  };

  constructor(
    private router: Router, 
    private passengerService: PassengerService,
    private busBookingService: BusBookingService,
    private loungeBookingService: LoungeBookingService,
    public notificationService: NotificationService,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    this.isBrowser = isPlatformBrowser(this.platformId);
    if (this.isBrowser) {
      Chart.register(...registerables);
    }
  }

  ngOnInit(): void {
    this.passengerService.passengers$.subscribe(ps => {
      this.passengers = ps;
      this.filteredPassengers = ps;
      this.refreshChart();
    });

    // Subscribe to bus bookings and update passenger service
    this.busBookingService.bookings$.subscribe(bookings => {
      this.passengerService.setBusBookingsData(bookings);
      this.refreshChart();
      this.refreshRouteChart();
    });

    // Subscribe to lounge bookings and update passenger service
    this.loungeBookingService.bookings$.subscribe(bookings => {
      this.passengerService.setLoungeBookingsData(bookings);
      this.refreshChart();
    });
  }

  goUserProfile() {
    this.router.navigate(['/user-profile']);
  }

  toggleNotificationPanel() {
    this.showNotificationPanel = !this.showNotificationPanel;
  }

  closeNotificationPanel() {
    this.showNotificationPanel = false;
  }

  showAddPassengerModal: boolean = false;
  newPassenger: Partial<Passenger> = {};

  showUpdatePassengerModal: boolean = false;
  selectedPassenger: Passenger | null = null;

  addPassenger(): void {
    this.showAddPassengerModal = true;
    this.newPassenger = {};
  }

  closeAddPassengerModal(): void {
    this.showAddPassengerModal = false;
    this.newPassenger = {};
  }

  savePassenger(): void {
    if (this.newPassenger.name && this.newPassenger.phone && this.newPassenger.email && this.newPassenger.nic) {
      this.passengerService.addPassenger(this.newPassenger as Passenger);
      this.closeAddPassengerModal();
    }
  }

  updatePassenger(p: Passenger): void {
    this.selectedPassenger = { ...p };
    this.showUpdatePassengerModal = true;
  }

  closeUpdatePassengerModal(): void {
    this.showUpdatePassengerModal = false;
    this.selectedPassenger = null;
  }

  saveUpdatedPassenger(): void {
    if (this.selectedPassenger && this.selectedPassenger.name && this.selectedPassenger.phone && this.selectedPassenger.email && this.selectedPassenger.nic) {
      this.passengerService.updatePassenger(this.selectedPassenger);
      this.closeUpdatePassengerModal();
    }
  }

  deletePassenger(p: Passenger): void {
    const ok = confirm(`Delete ${p.name}?`);
    if (ok) this.passengerService.deletePassenger(p.passenger_id);
  }

  onSearchChange(): void {
    const q = this.searchTerm.toLowerCase();
    this.filteredPassengers = !q ? this.passengers : this.passengers.filter(p =>
      p.name.toLowerCase().includes(q) ||
      p.email.toLowerCase().includes(q) ||
      p.phone.includes(q) ||
      p.nic.toLowerCase().includes(q) ||
      p.passenger_id.toLowerCase().includes(q)
    );
  }
  
  clearSearch(): void { this.searchTerm = ''; this.filteredPassengers = this.passengers; }

  refreshChart(): void { this.monthlyCounts = this.passengerService.getMonthlyCounts(new Date().getFullYear()); }
  
  refreshRouteChart(): void {
    const routeCounts = new Map<string, number>();
    this.busBookingService.bookings.forEach(booking => {
      if (booking.route) {
        routeCounts.set(booking.route, (routeCounts.get(booking.route) || 0) + 1);
      }
    });
    
    const totalCount = Array.from(routeCounts.values()).reduce((sum, count) => sum + count, 0);
    this.routeBookingCounts = Array.from(routeCounts.entries())
      .map(([route, count]) => ({
        route,
        count,
        percentage: totalCount > 0 ? count / totalCount : 0
      }))
      .sort((a, b) => b.count - a.count);
    
    // Update Chart.js data
    this.routeChartData = {
      labels: this.routeBookingCounts.map(item => item.route),
      datasets: [{
        data: this.routeBookingCounts.map(item => item.count),
        backgroundColor: ['#0046FF', '#9CA3AF', '#FB923C', '#60A5FA', '#0046FF', '#6B7280', '#FBBF24'],
        borderColor: ['#0046FF', '#9CA3AF', '#FB923C', '#60A5FA', '#0046FF', '#6B7280', '#FBBF24'],
        borderWidth: 1
      }]
    };
  }
    // 🚀 Export PDF Function
  exportHistoryPdf() {
    const doc = new jsPDF();
    doc.setFontSize(16);
    doc.text('Passenger History Report', 14, 15);

    autoTable(doc, {
      head: [['Passenger ID', 'Name', 'Phone', 'Email', 'NIC', 'Created']],
      body: this.filteredPassengers.map(p => [
        p.passenger_id,
        p.name,
        p.phone,
        p.email,
        p.nic,
        new Date(p.created_at).toISOString().split('T')[0]
      ]),
      startY: 25,
      theme: 'grid',
      headStyles: { fillColor: [59, 130, 246] }, // blue header
    });

    doc.save('Passenger_History.pdf');
  }


  // Helpers for inline SVG chart
  getMaxCount(): number { return Math.max(1, ...this.monthlyCounts); }
  getPoints(): string {
    const width = 600, height = 220, padding = 30;
    const max = this.getMaxCount();
    const stepX = (width - padding * 2) / (this.months.length - 1);
    return this.monthlyCounts
      .map((c, i) => {
        const x = padding + i * stepX;
        const y = height - padding - (c / max) * (height - padding * 2);
        return `${x},${y}`;
      })
      .join(' ');
  }

  toggleProfileMenu() {
    this.showProfileMenu = !this.showProfileMenu;
  }

  logout() {
    // Example logout logic
    localStorage.removeItem('token');
    this.router.navigate(['/login']);
  }

  onLogout() {
    this.logout();
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
    this.filteredPassengers = [...this.filteredPassengers].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (this.sortColumn) {
        case 'name':
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
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
