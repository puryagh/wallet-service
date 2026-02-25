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
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );

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
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    );