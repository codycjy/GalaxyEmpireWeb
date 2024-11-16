(globalThis.TURBOPACK = globalThis.TURBOPACK || []).push(["chunks/ssr/_e3292e._.js", {

"[project]/auth.config.ts [middleware] (ecmascript)": (({ r: __turbopack_require__, f: __turbopack_module_context__, i: __turbopack_import__, s: __turbopack_esm__, v: __turbopack_export_value__, n: __turbopack_export_namespace__, c: __turbopack_cache__, M: __turbopack_modules__, l: __turbopack_load__, j: __turbopack_dynamic__, P: __turbopack_resolve_absolute_path__, U: __turbopack_relative_url__, R: __turbopack_resolve_module_id_path__, g: global, __dirname, x: __turbopack_external_require__, y: __turbopack_external_import__, k: __turbopack_refresh__ }) => (() => {
"use strict";

__turbopack_esm__({
    "auth": ()=>auth,
    "authConfig": ()=>authConfig,
    "handlers": ()=>handlers,
    "signIn": ()=>signIn,
    "signOut": ()=>signOut
});
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$auth$2f$index$2e$js__$5b$middleware$5d$__$28$ecmascript$29$__$3c$module__evaluation$3e$__ = __turbopack_import__("[project]/node_modules/next-auth/index.js [middleware] (ecmascript) <module evaluation>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$auth$2f$index$2e$js__$5b$middleware$5d$__$28$ecmascript$29$__$3c$locals$3e$__ = __turbopack_import__("[project]/node_modules/next-auth/index.js [middleware] (ecmascript) <locals>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$auth$2f$providers$2f$credentials$2e$js__$5b$middleware$5d$__$28$ecmascript$29$__$3c$module__evaluation$3e$__ = __turbopack_import__("[project]/node_modules/next-auth/providers/credentials.js [middleware] (ecmascript) <module evaluation>");
var __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$auth$2f$core$2f$providers$2f$credentials$2e$js__$5b$middleware$5d$__$28$ecmascript$29$__ = __turbopack_import__("[project]/node_modules/@auth/core/providers/credentials.js [middleware] (ecmascript)");
"__TURBOPACK__ecmascript__hoisting__location__";
;
;
const authConfig = {
    providers: [
        (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f40$auth$2f$core$2f$providers$2f$credentials$2e$js__$5b$middleware$5d$__$28$ecmascript$29$__["default"])({
            credentials: {
                username: {
                    type: 'text',
                    label: 'Username'
                },
                password: {
                    type: 'password',
                    label: 'Password'
                },
                captcha: {
                    type: 'text',
                    label: 'Verification Code'
                }
            },
            async authorize (credentials) {
                try {
                    // ... 你的验证逻辑
                    return null;
                } catch (error) {
                    return null;
                }
            }
        })
    ],
    callbacks: {
        async jwt ({ token }) {
            return token;
        },
        async session ({ session }) {
            return session;
        }
    },
    pages: {
        signIn: '/'
    }
};
const { handlers, auth, signIn, signOut } = (0, __TURBOPACK__imported__module__$5b$project$5d2f$node_modules$2f$next$2d$auth$2f$index$2e$js__$5b$middleware$5d$__$28$ecmascript$29$__$3c$locals$3e$__["default"])(authConfig);

})()),
"[project]/middleware.ts [middleware] (ecmascript) <locals>": (({ r: __turbopack_require__, f: __turbopack_module_context__, i: __turbopack_import__, s: __turbopack_esm__, v: __turbopack_export_value__, n: __turbopack_export_namespace__, c: __turbopack_cache__, M: __turbopack_modules__, l: __turbopack_load__, j: __turbopack_dynamic__, P: __turbopack_resolve_absolute_path__, U: __turbopack_relative_url__, R: __turbopack_resolve_module_id_path__, g: global, __dirname, x: __turbopack_external_require__, y: __turbopack_external_import__, k: __turbopack_refresh__ }) => (() => {
"use strict";

// Protecting routes with next-auth
// https://next-auth.js.org/configuration/nextjs#middleware
// https://nextjs.org/docs/app/building-your-application/routing/middleware
__turbopack_esm__({
    "config": ()=>config
});
;
;
const config = {
    matcher: [
        '/((?!api|_next/static|_next/image|favicon.ico).*)'
    ]
};

})()),
"[project]/middleware.ts [middleware] (ecmascript) <module evaluation>": (({ r: __turbopack_require__, f: __turbopack_module_context__, i: __turbopack_import__, s: __turbopack_esm__, v: __turbopack_export_value__, n: __turbopack_export_namespace__, c: __turbopack_cache__, M: __turbopack_modules__, l: __turbopack_load__, j: __turbopack_dynamic__, P: __turbopack_resolve_absolute_path__, U: __turbopack_relative_url__, R: __turbopack_resolve_module_id_path__, g: global, __dirname, x: __turbopack_external_require__, y: __turbopack_external_import__ }) => (() => {
"use strict";

__turbopack_esm__({});
var __TURBOPACK__imported__module__$5b$project$5d2f$auth$2e$config$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__ = __turbopack_import__("[project]/auth.config.ts [middleware] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$locals$3e$__ = __turbopack_import__("[project]/middleware.ts [middleware] (ecmascript) <locals>");
"__TURBOPACK__ecmascript__hoisting__location__";

})()),
"[project]/middleware.ts [middleware] (ecmascript) <exports>": (({ r: __turbopack_require__, f: __turbopack_module_context__, i: __turbopack_import__, s: __turbopack_esm__, v: __turbopack_export_value__, n: __turbopack_export_namespace__, c: __turbopack_cache__, M: __turbopack_modules__, l: __turbopack_load__, j: __turbopack_dynamic__, P: __turbopack_resolve_absolute_path__, U: __turbopack_relative_url__, R: __turbopack_resolve_module_id_path__, g: global, __dirname, x: __turbopack_external_require__, y: __turbopack_external_import__ }) => (() => {
"use strict";

__turbopack_esm__({
    "config": ()=>__TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$locals$3e$__["config"],
    "middleware": ()=>__TURBOPACK__imported__module__$5b$project$5d2f$auth$2e$config$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__["auth"]
});
var __TURBOPACK__imported__module__$5b$project$5d2f$auth$2e$config$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__ = __turbopack_import__("[project]/auth.config.ts [middleware] (ecmascript)");
var __TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$locals$3e$__ = __turbopack_import__("[project]/middleware.ts [middleware] (ecmascript) <locals>");
"__TURBOPACK__ecmascript__hoisting__location__";

})()),
"[project]/middleware.ts [middleware] (ecmascript) <facade>": (({ r: __turbopack_require__, f: __turbopack_module_context__, i: __turbopack_import__, s: __turbopack_esm__, v: __turbopack_export_value__, n: __turbopack_export_namespace__, c: __turbopack_cache__, M: __turbopack_modules__, l: __turbopack_load__, j: __turbopack_dynamic__, P: __turbopack_resolve_absolute_path__, U: __turbopack_relative_url__, R: __turbopack_resolve_module_id_path__, g: global, __dirname, x: __turbopack_external_require__, y: __turbopack_external_import__ }) => (() => {
"use strict";

__turbopack_esm__({
    "config": ()=>__TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$exports$3e$__["config"],
    "middleware": ()=>__TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$exports$3e$__["middleware"]
});
var __TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$module__evaluation$3e$__ = __turbopack_import__("[project]/middleware.ts [middleware] (ecmascript) <module evaluation>");
var __TURBOPACK__imported__module__$5b$project$5d2f$middleware$2e$ts__$5b$middleware$5d$__$28$ecmascript$29$__$3c$exports$3e$__ = __turbopack_import__("[project]/middleware.ts [middleware] (ecmascript) <exports>");
"__TURBOPACK__ecmascript__hoisting__location__";

})()),
}]);

//# sourceMappingURL=_e3292e._.js.map