# Frontend Escalation Integration Guide

## Example: Adding Escalation Display to Complaint Cards

### 1. Import the Escalation Service

In your `complaint-management.component.ts`:

```typescript
import { EscalationService, ComplaintEscalation } from '../../core/services/escalation.service';

// Add to constructor
constructor(
  private complaintService: ComplaintService,
  private escalationService: EscalationService,
  private messageService: MessageService
) {}
```

### 2. Load Escalation Data with Complaints

```typescript
interface ComplaintWithEscalation extends Complaint {
  escalation?: ComplaintEscalation;
}

complaints: ComplaintWithEscalation[] = [];

loadComplaintsWithEscalation() {
  this.complaintService.loadComplaints().subscribe(complaints => {
    // Load escalation data for each complaint
    complaints.forEach(complaint => {
      this.escalationService.getComplaintEscalation(complaint.id).subscribe({
        next: (escalation) => {
          const complaintWithEscalation = complaint as ComplaintWithEscalation;
          complaintWithEscalation.escalation = escalation;
        },
        error: (err) => {
          // No escalation data yet - that's okay
          console.log(`No escalation for complaint ${complaint.id}`);
        }
      });
    });
    this.complaints = complaints;
  });
}
```

### 3. Add HTML Template for Escalation Badge

Add this to your complaint card template (e.g., in the header section):

```html
<!-- Escalation Badge -->
<div class="escalation-info" *ngIf="complaint.escalation">
  <div class="escalation-badge" 
       [ngClass]="'level-' + complaint.escalation.current_level">
    <i class="pi pi-shield"></i>
    <span>L{{ complaint.escalation.current_level }}: {{ complaint.escalation.current_team }}</span>
  </div>
  
  <div class="escalation-timer" 
       *ngIf="complaint.escalation.next_escalation_due"
       [ngClass]="{'overdue': isEscalationOverdue(complaint.escalation.next_escalation_due)}">
    <i class="pi pi-clock"></i>
    <span *ngIf="!isEscalationOverdue(complaint.escalation.next_escalation_due)">
      Escalates in {{ getDaysUntilEscalation(complaint.escalation.next_escalation_due) }} days
    </span>
    <span *ngIf="isEscalationOverdue(complaint.escalation.next_escalation_due)" class="overdue-text">
      Overdue for escalation!
    </span>
  </div>
</div>
```

### 4. Add Component Methods

```typescript
getDaysUntilEscalation(nextEscalationDue: string): number {
  return this.escalationService.getDaysUntilEscalation(nextEscalationDue);
}

isEscalationOverdue(nextEscalationDue: string): boolean {
  return this.escalationService.isOverdue(nextEscalationDue);
}

manualEscalate(complaint: ComplaintWithEscalation) {
  if (!complaint.escalation) {
    this.messageService.add({
      severity: 'error',
      summary: 'Error',
      detail: 'No escalation data available'
    });
    return;
  }

  this.escalationService.escalateComplaint(
    complaint.id,
    this.mapCategoryToIssueType(complaint.category),
    complaint.escalation.current_level
  ).subscribe({
    next: () => {
      this.messageService.add({
        severity: 'success',
        summary: 'Success',
        detail: 'Complaint escalated to next level'
      });
      this.loadComplaintsWithEscalation();
    },
    error: (err) => {
      this.messageService.add({
        severity: 'error',
        summary: 'Error',
        detail: 'Failed to escalate complaint'
      });
    }
  });
}

// Helper to map display category to database issue_type
mapCategoryToIssueType(category: string): string {
  const mapping: { [key: string]: string } = {
    'Operations & Scheduling': 'bus_delay',
    'Vehicle & Facility': 'flat_wheel', // or maintenance_issue
    'Service Issue': 'passenger_complaint',
    'Safety & Security': 'safety_concern'
  };
  return mapping[category] || 'other';
}
```

### 5. Add Styles (in component.scss)

```scss
.escalation-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 0.5rem;

  .escalation-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.75rem;
    border-radius: 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    width: fit-content;

    &.level-1 {
      background: #e3f2fd;
      color: #1976d2;
      border: 1px solid #1976d2;
    }

    &.level-2 {
      background: #fff3e0;
      color: #f57c00;
      border: 1px solid #f57c00;
    }

    &.level-3 {
      background: #ffebee;
      color: #d32f2f;
      border: 1px solid #d32f2f;
    }
  }

  .escalation-timer {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.875rem;
    color: #666;

    &.overdue {
      color: #d32f2f;
      font-weight: 600;

      .overdue-text {
        animation: pulse 2s infinite;
      }
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}
```

### 6. Add Manual Escalation Button

Add this button to your complaint card actions:

```html
<button 
  pButton 
  label="Escalate" 
  icon="pi pi-arrow-up"
  class="p-button-warning p-button-sm"
  *ngIf="complaint.escalation && complaint.escalation.current_level < 3"
  (click)="manualEscalate(complaint)"
  [disabled]="complaint.status === 'Resolved'">
</button>
```

### 7. Show Escalation History Dialog

Add a button to view history:

```html
<button 
  pButton 
  label="View History" 
  icon="pi pi-history"
  class="p-button-text p-button-sm"
  *ngIf="complaint.escalation"
  (click)="showEscalationHistory(complaint.id)">
</button>
```

Component method:

```typescript
showEscalationHistory(complaintId: string) {
  this.escalationService.getEscalationHistory(complaintId).subscribe({
    next: (response) => {
      // Show in a dialog or sidebar
      console.log('Escalation History:', response.history);
      // You can use PrimeNG Dialog to display this
    }
  });
}
```

### 8. Initialize Escalation for New Complaints

When a new complaint is created, initialize escalation automatically:

```typescript
onComplaintCreated(complaint: Complaint) {
  // Initialize escalation
  const issueType = this.mapCategoryToIssueType(complaint.category);
  
  this.escalationService.initializeEscalation(complaint.id, issueType).subscribe({
    next: () => {
      console.log('Escalation initialized for complaint:', complaint.id);
      this.loadComplaintsWithEscalation();
    },
    error: (err) => {
      console.error('Failed to initialize escalation:', err);
    }
  });
}
```

### 9. Display Escalation Stats Dashboard Widget

Create a dashboard widget showing escalation stats:

```typescript
escalationStats: EscalationStats | null = null;

loadEscalationStats() {
  this.escalationService.getEscalationStats().subscribe(stats => {
    this.escalationStats = stats;
  });
}
```

Template:

```html
<div class="escalation-stats-card">
  <h3>Escalation Overview</h3>
  
  <div class="stats-grid" *ngIf="escalationStats">
    <div class="stat-item level-1">
      <div class="stat-value">{{ escalationStats.by_level[1] || 0 }}</div>
      <div class="stat-label">Level 1</div>
    </div>
    
    <div class="stat-item level-2">
      <div class="stat-value">{{ escalationStats.by_level[2] || 0 }}</div>
      <div class="stat-label">Level 2</div>
    </div>
    
    <div class="stat-item level-3">
      <div class="stat-value">{{ escalationStats.by_level[3] || 0 }}</div>
      <div class="stat-label">Level 3</div>
    </div>
    
    <div class="stat-item overdue" *ngIf="escalationStats.overdue_count > 0">
      <div class="stat-value">{{ escalationStats.overdue_count }}</div>
      <div class="stat-label">Overdue</div>
    </div>
  </div>
</div>
```

## Complete Example Component

Here's a complete example showing all features:

```typescript
import { Component, OnInit } from '@angular/core';
import { EscalationService, ComplaintEscalation, EscalationStats } from '../../core/services/escalation.service';
import { ComplaintService } from '../../core/services/complaint.service';

@Component({
  selector: 'app-complaint-management',
  templateUrl: './complaint-management.component.html',
  styleUrls: ['./complaint-management.component.scss']
})
export class ComplaintManagementComponent implements OnInit {
  complaints: any[] = [];
  escalationStats: EscalationStats | null = null;
  
  constructor(
    private complaintService: ComplaintService,
    private escalationService: EscalationService,
    private messageService: MessageService
  ) {}
  
  ngOnInit() {
    this.loadData();
  }
  
  loadData() {
    // Load complaints with escalation data
    this.complaintService.loadComplaints().subscribe(complaints => {
      this.complaints = complaints;
      
      // Load escalation for each complaint
      complaints.forEach(complaint => {
        this.escalationService.getComplaintEscalation(complaint.id).subscribe({
          next: (escalation) => {
            complaint.escalation = escalation;
          },
          error: () => {
            // No escalation yet
          }
        });
      });
    });
    
    // Load stats
    this.loadEscalationStats();
  }
  
  loadEscalationStats() {
    this.escalationService.getEscalationStats().subscribe(stats => {
      this.escalationStats = stats;
    });
  }
  
  getDaysUntilEscalation(nextEscalationDue: string): number {
    return this.escalationService.getDaysUntilEscalation(nextEscalationDue);
  }
  
  isEscalationOverdue(nextEscalationDue: string): boolean {
    return this.escalationService.isOverdue(nextEscalationDue);
  }
  
  manualEscalate(complaint: any) {
    this.escalationService.escalateComplaint(
      complaint.id,
      this.mapCategoryToIssueType(complaint.category),
      complaint.escalation.current_level
    ).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Success',
          detail: 'Complaint escalated successfully'
        });
        this.loadData();
      },
      error: () => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Failed to escalate complaint'
        });
      }
    });
  }
  
  mapCategoryToIssueType(category: string): string {
    const mapping: { [key: string]: string } = {
      'Operations & Scheduling': 'bus_delay',
      'Vehicle & Facility': 'flat_wheel',
      'Service Issue': 'passenger_complaint',
      'Safety & Security': 'safety_concern'
    };
    return mapping[category] || 'other';
  }
}
```

## Testing the Integration

1. **Open your browser** to http://localhost:4200
2. **Navigate to Complaint Management** page
3. **You should see:**
   - Escalation badges showing the current level and team
   - Timer showing days until escalation
   - Manual escalation buttons
   - Escalation statistics dashboard

## Important Notes

- Escalation is **automatically initialized** when backend receives a new complaint
- The **scheduler runs every hour** to check for overdue complaints
- Complaints at **Level 3** cannot be escalated further
- **Resolved/Closed** complaints are excluded from automatic escalation
