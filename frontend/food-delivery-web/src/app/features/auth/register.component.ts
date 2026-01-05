import { Component, inject, signal } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { Router } from '@angular/router';
import { AuthenticationService } from '../../api';
import { AuthService } from '../../core/auth/auth.service';

@Component({
  selector: 'fdp-register',
  standalone: true,
  imports: [ReactiveFormsModule, MatFormFieldModule, MatInputModule, MatButtonModule],
  template: `
    <div style="max-width:420px;margin:48px auto;padding:24px;">
      <h2>Register</h2>
      <form [formGroup]="form" (ngSubmit)="submit()" style="display:grid; gap:16px;">
        <mat-form-field appearance="outline">
          <mat-label>Email</mat-label>
          <input matInput formControlName="email" type="email" required />
        </mat-form-field>
        <mat-form-field appearance="outline">
          <mat-label>Password</mat-label>
          <input matInput formControlName="password" type="password" required />
        </mat-form-field>
        <button mat-raised-button color="primary" type="submit" [disabled]="form.invalid || loading()">Create account</button>
        <div *ngIf="error()" style="color:#b00020;">{{ error() }}</div>
      </form>
    </div>
  `,
})
export class RegisterComponent {
  private fb = inject(FormBuilder);
  private api = inject(AuthenticationService);
  private auth = inject(AuthService);
  private router = inject(Router);

  loading = signal(false);
  error = signal<string | null>(null);

  form = this.fb.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(6)]],
  });

  submit() {
    if (this.form.invalid) return;
    this.loading.set(true);
    this.error.set(null);
    const body = {
      email: this.form.value.email!,
      password: this.form.value.password!,
    };
    this.api.authRegisterPost(body).subscribe({
      next: (res) => {
        const token = (res as any)?.token;
        if (token) this.auth.setToken(token);
        this.router.navigateByUrl('/');
      },
      error: (err) => {
        this.error.set(err?.error?.error || 'Registration failed');
        this.loading.set(false);
      },
    });
  }
}
