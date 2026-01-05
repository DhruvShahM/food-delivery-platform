import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { OrdersService } from '../../api';

@Component({
  selector: 'fdp-orders',
  standalone: true,
  imports: [CommonModule, MatListModule, MatProgressSpinnerModule],
  template: `
    <div style="padding:24px">
      <h2>My Orders</h2>
      <ng-container *ngIf="loading(); else listTpl">
        <div style="display:flex;justify-content:center;margin-top:24px;">
          <mat-spinner diameter="36"></mat-spinner>
        </div>
      </ng-container>
      <ng-template #listTpl>
        <mat-nav-list *ngIf="orders()?.orders?.length; else emptyTpl">
          <a mat-list-item *ngFor="let o of orders()!.orders">
            <div matListItemTitle>Order {{ o.id }} - {{ o.status }}</div>
            <div matListItemLine>Amount: {{ o.amount | number:'1.2-2' }}</div>
          </a>
        </mat-nav-list>
        <ng-template #emptyTpl>
          <p>No orders yet.</p>
        </ng-template>
      </ng-template>
    </div>
  `,
})
export class OrdersComponent implements OnInit {
  private api = inject(OrdersService);
  loading = signal(true);
  orders = signal<{ orders?: any[] } | null>(null);

  ngOnInit(): void {
    this.api.apiV1OrdersGet().subscribe({
      next: (res) => {
        this.orders.set(res as any);
        this.loading.set(false);
      },
      error: () => {
        this.orders.set({ orders: [] });
        this.loading.set(false);
      },
    });
  }
}
