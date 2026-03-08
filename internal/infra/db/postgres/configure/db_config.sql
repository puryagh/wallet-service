-- initialize base data

-- insert IRR asset
INSERT INTO
    wallet_assets (
        active,
        code,
        symbol,
        title,
        description,
        decimals,
        network,
        icon_url,
        unit,
        unit_title,
        meta_data,
        ledger_code,
        predefined,
        created_at,
        updated_at
    )
VALUES
    (
        true,
        '1',
        'IRR',
        'ریال',
        'ریال ایران',
        0,
        'IRCBN',
        'https://www.noghrestan.com/assets/icons/irr.svg',
        'PIECE',
        'ریال',
        '{
            "allowed_pairs": "IRR_USD,IRR_XAG_MILLIGRAM,IRR_XAG_GRAM"
        }',
        1,
        True,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

-- insert default issuer bin for secure wallet cards
INSERT INTO
    issuer_bins (
        active,
        bin,
        brand,
        issuer_name,
        country_code,
        meta_data,
        created_at,
        updated_at
    )
VALUES
    (
        true,
        '502229',
        'LOCAL',
        'LIVEUTIL',
        'IR',
        '{
            "managed_by": "wallet-service",
            "seed_source": "db_config"
        }',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
ON CONFLICT (bin) DO UPDATE
SET
    active = EXCLUDED.active,
    brand = EXCLUDED.brand,
    issuer_name = EXCLUDED.issuer_name,
    country_code = EXCLUDED.country_code,
    meta_data = EXCLUDED.meta_data,
    updated_at = CURRENT_TIMESTAMP;

-- insert default magnetic card product bound to the default issuer
INSERT INTO
    card_products (
        issuer_bin_id,
        name,
        pan_length,
        cvv_length,
        expiry_months,
        service_code,
        allow_magnetic_stripe,
        meta_data,
        created_at,
        updated_at
    )
SELECT
    issuer_bins.id,
    'DEFAULT_MAGNETIC_CARD',
    16,
    3,
    36,
    '201',
    true,
    '{
        "managed_by": "wallet-service",
        "seed_source": "db_config"
    }',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM issuer_bins
WHERE issuer_bins.bin = '502229'
ON CONFLICT (name) DO UPDATE
SET
    issuer_bin_id = EXCLUDED.issuer_bin_id,
    pan_length = EXCLUDED.pan_length,
    cvv_length = EXCLUDED.cvv_length,
    expiry_months = EXCLUDED.expiry_months,
    service_code = EXCLUDED.service_code,
    allow_magnetic_stripe = EXCLUDED.allow_magnetic_stripe,
    meta_data = EXCLUDED.meta_data,
    updated_at = CURRENT_TIMESTAMP;

-- insert XAG asset
INSERT INTO
    wallet_assets (
        active,
        code,
        symbol,
        title,
        description,
        decimals,
        network,
        icon_url,
        unit,
        unit_title,
        meta_data,
        ledger_code,
        predefined,
        created_at,
        updated_at
    )
VALUES
    (
        true,
        '2',
        'XAG',
        'نقره',
        'میلی‌گرم نقره',
        2,
        'LIVEUTIL',
        'https://www.noghrestan.com/assets/icons/xag.svg',
        'MILLIGRAM',
        'میلی‌گرم',
        '{
            "allowed_pairs": "XAG_MILLIGRAM_IRR,XAG_MILLIGRAM_USD,XAG_GRAM_IRR,XAG_GRAM_USD,XAG_OZ_USD"
        }',
        2,
        True,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );