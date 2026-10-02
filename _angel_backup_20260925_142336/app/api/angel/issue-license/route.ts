import { NextResponse } from 'next/server';
import crypto from 'crypto';

type IssueRequest = {
  customer_id?: string;
  branch_id?: string;
  industry_id?: string;
  hardware_id?: string;
  order_id?: string;
};

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as IssueRequest;

    if (
      !body.customer_id ||
      !body.branch_id ||
      !body.industry_id ||
      !body.hardware_id ||
      !body.order_id
    ) {
      return NextResponse.json(
        {
          error: 'missing_required_fields',
          message: 'تمام فیلدهای صدور لایسنس الزامی هستند.',
        },
        { status: 400 }
      );
    }

    const licenseId =
      'HZ-' +
      crypto.randomBytes(8).toString('hex').toUpperCase();

    /*
     * DEMO ONLY
     *
     * این endpoint فعلاً UI flow را تست می‌کند.
     * private key نباید در browser یا این route آزمایشی قرار گیرد.
     *
     * در اتصال واقعی:
     *   customer_id
     *   branch_id
     *   industry_id
     *   hardware_id
     *   order_id
     *
     * باید به سرویس صدور واقعی Horizon Core منتقل شوند.
     */

    return NextResponse.json({
      ok: true,
      license_id: licenseId,
      customer_id: body.customer_id,
      branch_id: body.branch_id,
      industry_id: body.industry_id,
      hardware_id: body.hardware_id,
      order_id: body.order_id,
      mode: 'demo',
    });
  } catch {
    return NextResponse.json(
      {
        error: 'invalid_request',
        message: 'JSON درخواست معتبر نیست.',
      },
      { status: 400 }
    );
  }
}
