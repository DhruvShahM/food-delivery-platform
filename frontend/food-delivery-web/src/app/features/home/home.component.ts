import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';

@Component({
  selector: 'fdp-home',
  standalone: true,
  imports: [RouterLink, MatButtonModule],
  template: `
    <div style="padding:24px">
      <h1>Food Delivery</h1>
      <p>Welcome! Please login or register to continue.</p>
      <div style="display:flex; gap:12px; margin-top:16px;">
        <a mat-raised-button color="primary" routerLink="/login">Login</a>
        <a mat-raised-button color="accent" routerLink="/register">Register</a>
      </div>
    </div>
  `,
})
export class HomeComponent {}
