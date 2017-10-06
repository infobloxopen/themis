/*--------------------------------------------------------------------------*\
 *
 * TrustedSource SDK
 *
 * Copyright (C) 2009 McAfee, Inc. All Rights Reserved.
 *
 * $RCSfile: ts.h,v $
 * $Revision: 1174 $
 * $Date: 2015-10-06 03:40:37 -0700 (Tue, 06 Oct 2015) $
 * $State: Exp $
 * $Author: cdhulipa $
 \*--------------------------------------------------------------------------*/

#ifndef TS_WEB_H
#define TS_WEB_H

#include <stdio.h>
#include <sys/types.h>

#if defined(WIN32) || defined(WIN64)
/* NOTE: does not have an effect once windows.h has been included */
#   ifndef WIN32_LEAN_AND_MEAN
#       define WIN32_LEAN_AND_MEAN  1
#   endif /* WIN32_LEAN_AND_MEAN */
#   include <windows.h>
#else
#   include <sys/socket.h>
#   include <netinet/in.h>
#   include <netdb.h>
#endif

#ifdef __cplusplus
extern "C" {
#endif /* __cplusplus */

/* general compilation features */
#ifndef TS_FEATURE_FUNCTION_DECLARATIONS
#   define TS_FEATURE_FUNCTION_DECLARATIONS     1
#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                        GENERAL
 *
 \*--------------------------------------------------------------*/
#if defined(WIN32) || defined(WIN64)
# ifdef TS_API_BUILD
#  define TS_API_PUBLIC   __declspec(dllexport)
# else /* TS_API_BUILD */
#  ifdef TS_API_STATIC
#   define TS_API_PUBLIC
#  else
#   define TS_API_PUBLIC   __declspec(dllimport)
#  endif
# endif /* TS_API_BUILD */
# ifndef __attribute__
#  define __attribute__(attr)
# endif
#else
# define TS_API_PUBLIC   extern
# if defined (__x86_64__) || defined(SOLARIS) || defined(PPC)
#  define _cdecl
# else
#  define _cdecl __attribute__ ((cdecl))
# endif
#endif /* WIN32 || WIN64 */

/*
 * We define timeout as long long on Win64 and as long on all other
 * platforms to maintain ABI compatibility.
 */
#undef TS_USE_LONG_LONG_TIMEOUT
#if defined(WIN64)
#   define TS_USE_LONG_LONG_TIMEOUT 1
#endif

#define TS_API_INTERNAL  extern
#define TS_API_PRIVATE   static

#ifdef TS_UNITTEST
#   define TS_API_PRIVATEUT /* nothing */
#else
#   define TS_API_PRIVATEUT TS_API_PRIVATE
#endif

#define TS_API_PROTECTED TS_API_PUBLIC

#define TS_API_DESC_LEN             80
#define TS_API_VERSION_LEN          20

/*
 * Return values from functions
 */
#define TS_OK                  0
#define TS_ERROR               1

#define TS_NOMEM              10
#define TS_INSUFFICIENT_SPACE 11

#define TS_INVALID_URL        20
#define TS_URL_TOO_LONG       21

#define TS_INVALID_DATABASE   30
#define TS_NO_DATABASE        31
#define TS_EXPIRED_DATABASE   32

#define TS_INVALID_FILE       40
#define TS_DOWNLOAD_FAILED    41
#define TS_MERGE_FAILED       42
#define TS_DOWNLOAD_CANCELLED 43

#define TS_PENDING            50
#define TS_TIMEOUT            51
#define TS_NOTFOUND           52
#define TS_ASYNC              53
#define TS_SYNC               54
#define TS_THROTTLE           55

#define TS_USE_DEFAULT        -1

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_Init(void);

TS_API_PUBLIC int _cdecl
TS_Shutdown(void);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */


/*
 * Values to turn features on and off
 */
#define TS_ENABLE  1
#define TS_DISABLE 2


/*--------------------------------------------------------------*\
 *
 *                          HANDLE
 *
 \*--------------------------------------------------------------*/
typedef struct ts_handle *TS_Handle;

#if TS_FEATURE_FUNCTION_DECLARATIONS

/* [ts_connection_sharing] allowed */
TS_API_PUBLIC int _cdecl
TS_HandleCreate(TS_Handle *ts_handle,
                const char *serial_number,
                const char *client_id,
                const char *product_type,
                const char *product_version);

/* [ts_connection_sharing] allowed */
TS_API_PUBLIC int _cdecl
TS_HandleDestroy(TS_Handle *ts_handle);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/* [ts_connection_sharing] {{{ */
#define TS_HANDLE_TYPE_STANDARD                     1
#define TS_HANDLE_TYPE_IPCSERVER                    2
/* }}} */

typedef enum
{
    TS_HANDLE_INFO_FIRST,  /* do not use */

    TS_HANDLE_INFO_DATABASE_EXPIRED,
    TS_HANDLE_INFO_DATABASE_EXPIRE_TIME,
    TS_HANDLE_INFO_DATABASE_SERIAL_NUM,
    TS_HANDLE_INFO_DATABASE_MAX_CATSET,
    TS_HANDLE_INFO_API_VERSION_MAJOR,
    TS_HANDLE_INFO_API_VERSION_MINOR,
    TS_HANDLE_INFO_API_VERSION_DESC,
    TS_HANDLE_INFO_API_VERSION_STR,
    TS_HANDLE_INFO_HITCOUNT_UNRATED,
    TS_HANDLE_INFO_HITCOUNT_DATABASE,
    TS_HANDLE_INFO_HITCOUNT_CUSTOMSITES,
    TS_HANDLE_INFO_HITCOUNT_TSCACHE,
    TS_HANDLE_INFO_HITCOUNT_PATTERNS,
    TS_HANDLE_INFO_HITCOUNT_TRUSTEDSOURCE,
    TS_HANDLE_INFO_AVGTIME_LOCALLY_RATED,
    TS_HANDLE_INFO_AVGTIME_LOCALLY_UNRATED,
    TS_HANDLE_INFO_AVGTIME_NETWORK_RATED,
    TS_HANDLE_INFO_AVGTIME_NETWORK_UNRATED,
    TS_HANDLE_INFO_HANDLE_QUERYTYPE, /* info type: unsigned int. values: TS_HANDLE_TYPE_* */
    TS_HANDLE_INFO_URLCACHE_DIRECT_HITS,
    TS_HANDLE_INFO_URLCACHE_INDIRECT_HITS,

    TS_HANDLE_INFO_LAST  /* do not use */
} TS_Handle_Info;

#if TS_FEATURE_FUNCTION_DECLARATIONS

/* [ts_connection_sharing] allowed */
TS_API_PUBLIC int _cdecl
TS_HandleInfoGet(TS_Handle ts_handle,
                 TS_Handle_Info info_type,
                 void *info,
                 unsigned int info_size);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */


/*--------------------------------------------------------------*\
 *
 *                       ATTRIBUTES
 *
 \*--------------------------------------------------------------*/
#define TS_REPUTATION_NEUTRAL 0
#define TS_REPUTATION_TRUSTED -127

typedef struct ts_attributes *TS_Attributes;

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_AttributesCreate(TS_Handle ts_handle,
                    TS_Attributes *attributes);

TS_API_PUBLIC int _cdecl
TS_AttributesDestroy(TS_Handle ts_handle,
                     TS_Attributes *attributes);

TS_API_PUBLIC int _cdecl
TS_AttributesReset(TS_Handle ts_handle,
                   TS_Attributes attributes);

TS_API_PUBLIC int _cdecl
TS_AttributesCopy(TS_Handle ts_handle,
                  TS_Attributes src,
                  TS_Attributes dst);

TS_API_PUBLIC int
TS_AttributesAddTSQueryField(TS_Handle ts_handle,
                             TS_Attributes attributes,
                             const unsigned int field_id,
                             const void* field_value,
                             const unsigned int field_value_size);

TS_API_PUBLIC int
TS_AttributesGetTSResponseField(TS_Handle ts_handle,
                                TS_Attributes attributes,
                                const unsigned int field_id,
                                void* field_value,
                                unsigned int *field_value_size);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum
{
    TS_ATTRIBUTES_INFO_FIRST,

    TS_ATTRIBUTES_INFO_REPUTATION,
    TS_ATTRIBUTES_INFO_SRC_REPUTATION,
    TS_ATTRIBUTES_INFO_DEST_REPUTATION,
    TS_ATTRIBUTES_INFO_GEOLOCATION,
    TS_ATTRIBUTES_INFO_SRC_GEOLOCATION,
    TS_ATTRIBUTES_INFO_DEST_GEOLOCATION,
    TS_ATTRIBUTES_INFO_SRC_TTL,
    TS_ATTRIBUTES_INFO_DEST_TTL,
    TS_ATTRIBUTES_INFO_REPUTATION_STATUS,
    TS_ATTRIBUTES_INFO_DYNAMIC_QUARANTINE,
    TS_ATTRIBUTES_INFO_RAW_REQUEST,
    TS_ATTRIBUTES_INFO_RAW_REQUEST_LEN,
    TS_ATTRIBUTES_INFO_MCAFEE_SECURE,

    TS_ATTRIBUTES_INFO_LAST
} TS_Attributes_Info;

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int
TS_AttributesInfoGet(TS_Handle          ts_handle,
                     TS_Attributes      attributes,
                     TS_Attributes_Info info_type,
                     void              *info,
                     unsigned int       info_size);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                       CATEGORIES
 *
 \*--------------------------------------------------------------*/
typedef struct ts_categories *TS_Categories;

typedef enum
{
    TS_LANGUAGE_FIRST, /* do not use */

    TS_LANGUAGE_ENGLISH,
    TS_LANGUAGE_ENGLISH_CATEGORY_SHORT,
    TS_LANGUAGE_GERMAN,
    TS_LANGUAGE_FRENCH,
    TS_LANGUAGE_SPANISH,
    TS_LANGUAGE_PORTUGUESE,
    TS_LANGUAGE_RUSSIAN,
    TS_LANGUAGE_ITALIAN,
    TS_LANGUAGE_KOREAN,
    TS_LANGUAGE_JAPANESE,
    TS_LANGUAGE_CHINESE_SIMPLIFIED,
    TS_LANGUAGE_CHINESE_TRADITIONAL,

    TS_LANGUAGE_LAST /* do not use */
} TS_Language;

typedef enum
{
    TS_ENCODING_FIRST, /* do not use */

    TS_ENCODING_UTF8,

    TS_ENCODING_LAST /* do not use */
} TS_Encoding;

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_CategoriesCategoryAdd(TS_Handle ts_handle,
                         TS_Categories categories,
                         unsigned int category);

TS_API_PUBLIC int _cdecl
TS_CategoriesCategoryAddAll(TS_Handle ts_handle,
                            TS_Categories categories,
                            unsigned int category_set_version);

TS_API_PUBLIC int _cdecl
TS_CategoriesCategoryIsMember(TS_Handle ts_handle,
                              TS_Categories categories,
                              unsigned int category,
                              int *member);

TS_API_PUBLIC int _cdecl
TS_CategoriesCategoryRemove(TS_Handle ts_handle,
                            TS_Categories categories,
                            unsigned int category);

TS_API_PUBLIC int _cdecl
TS_CategoriesCategoryRemoveAll(TS_Handle ts_handle,
                               TS_Categories categories);

TS_API_PUBLIC int _cdecl
TS_CategoriesCopy(TS_Handle ts_handle,
                  TS_Categories src,
                  TS_Categories dst);

TS_API_PUBLIC int _cdecl
TS_CategoriesCount(TS_Handle ts_handle,
                   TS_Categories categories,
                   int *count);

TS_API_PUBLIC int _cdecl
TS_CategoriesCreate(TS_Handle ts_handle,
                    TS_Categories *categories);

TS_API_PUBLIC int _cdecl
TS_CategoriesDestroy(TS_Handle ts_handle,
                     TS_Categories *categories);

TS_API_PUBLIC int _cdecl
TS_CategoriesEqual(TS_Handle ts_handle,
                   TS_Categories categories1,
                   TS_Categories categories2,
                   int *equal);

TS_API_PUBLIC int _cdecl
TS_CategoriesIntersect(TS_Handle ts_handle,
                       TS_Categories categories1,
                       TS_Categories categories2,
                       TS_Categories result,
                       int *num_cats);

TS_API_PUBLIC int _cdecl
TS_CategoriesToArray(TS_Handle ts_handle,
                     TS_Categories categories,
                     unsigned int *cat_array,
                     int *num_cats);

TS_API_PUBLIC int _cdecl
TS_CategoriesToString(TS_Handle ts_handle,
                      TS_Categories categories,
                      TS_Language language,
                      TS_Encoding encoding,
                      char *delimiter,
                      int delimiter_len,
                      char *cat_string,
                      int *cat_string_len);

TS_API_PUBLIC int _cdecl
TS_CategoriesUnion(TS_Handle ts_handle,
                   TS_Categories categories1,
                   TS_Categories categories2,
                   TS_Categories result,
                   int *num_cats);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                  RATING SESSION
 *
 \*--------------------------------------------------------------*/
typedef struct ts_rating_session *TS_RatingSession;


/*--------------------------------------------------------------*\
 *
 *                       URLS
 *
 \*--------------------------------------------------------------*/
typedef struct ts_url *TS_Url;

#define TS_RATE_URL_SKIP_CUSTOM_KEYWORDS   0x00000001
#define TS_RATE_URL_SKIP_CUSTOM_SITES      0x00000002
#define TS_RATE_URL_SKIP_DATABASE          0x00000004
#define TS_RATE_URL_SKIP_DATABASE_KEYWORDS 0x00000008
#define TS_RATE_URL_SKIP_PATTERNS          0x00000010
#define TS_RATE_URL_SKIP_CGIS              0x00000020
#define TS_RATE_URL_SKIP_EMBEDDED_URLS     0x00000040
#define TS_RATE_URL_SKIP_TRUSTEDSOURCE     0x00000080
#define TS_RATE_URL_SKIP_DNS               0x00000100
#define TS_RATE_URL_SKIP_WEBREP_PRESERVE   0x00000200
#define TS_RATE_SKIP_CACHE                 0x00000400

#define TS_RATE_URL_SKIP_MASK              0x000007FF
#define TS_RATE_URL_SKIP_NONE              (0xFFFFFFFF ^ TS_RATE_URL_SKIP_MASK)
#define TS_RATE_URL_SKIP_ALL               (0xFFFFFFFF & TS_RATE_URL_SKIP_MASK)

/*
 * These are defined here for placeholders for internally used
 * flags
 *
#define TS_RATE_URL_FLAG_INTERNALUSE1            0x00010000
#define TS_RATE_URL_FLAG_INTERNALUSE2            0x00040000
#define TS_RATE_URL_FLAG_INTERNALUSE3            0x00080000
#define TS_RATE_URL_FLAG_INTERNALUSE4            0x00200000
 */
#define TS_RATE_FLAG_SYNC                        0x00020000
#define TS_RATE_URL_FLAG_UNCAT_DOMAIN_IP_LOOKUP  0x00100000
#define TS_RATE_MESSAGE_REQUERY                  0x00400000
#define TS_RATE_FLAG_TEST_QUERY                  0x00800000
#define TS_RATE_IPSPAM_REQUERY                   0x01000000
#define TS_RATE_FLAG_TS_FIELD_RESP               0x02000000

#define TS_RATE_FLAG_CONN_INBOUND                0x04000000
#define TS_RATE_FLAG_CONN_OUTBOUND               0x08000000

#define TS_RATE_URL_FLAG_MASK                    0x03FF0000
#define TS_RATE_URL_FLAG_ALL                     TS_RATE_URL_FLAG_MASK

typedef enum
{
    TS_RATING_URL           = 1,
    TS_RATING_MESSAGE,
    TS_RATING_CONNECTION,
    TS_RATING_IPSPAM,
    TS_RATING_MAX
} TS_RatingSessionType;

typedef void (*TS_RatingCompleteFunc)(TS_Handle ts_handle,
                                      TS_RatingSession session,
                                      TS_RatingSessionType type,
                                      void *userdata);

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_RateUrlExtended(TS_Handle ts_handle,
                   TS_Url url,
                   TS_Attributes attributes,
                   TS_Categories categories,
                   int *num_cats,
                   int flags,
                   unsigned int category_set,
                   unsigned int session_id,
                   char *user_agent,
                   TS_RatingSession *session,
                   TS_RatingCompleteFunc callback,
                   void *callback_data,
                   struct in6_addr *ips,
                   unsigned int ip_count);

TS_API_PUBLIC int _cdecl
TS_RateUrl(TS_Handle ts_handle,
           TS_Url url,
           TS_Attributes attributes,
           TS_Categories categories,
           int *num_cats,
           int flags,
           unsigned int category_set,
           unsigned int session_id,
           char *user_agent);

TS_API_PUBLIC int _cdecl
TS_UrlCreate(TS_Handle ts_handle,
             TS_Url *url);

TS_API_PUBLIC int _cdecl
TS_UrlDestroy(TS_Handle ts_handle,
              TS_Url *url);

TS_API_PUBLIC int _cdecl
TS_UrlDomainRewrite(TS_Handle ts_handle,
                    TS_Url url,
                    const char *new_domain);

TS_API_PUBLIC int _cdecl
TS_UrlCopy(TS_Handle ts_handle,
           TS_Url src,
           TS_Url dst);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum
{
    TS_URL_INFO_FIRST, /* do not use */

    TS_URL_INFO_DOMAIN,
    TS_URL_INFO_FILE_EXT,
    TS_URL_INFO_NUM_PATHS,
    TS_URL_INFO_DOMAIN_IS_IP,
    TS_URL_INFO_IS_USER_PAGE,
    TS_URL_INFO_PROTOCOL,
    TS_URL_INFO_PORT,
    TS_URL_INFO_PATH,
    TS_URL_INFO_CGI,
    TS_URL_INFO_CACHE_PATH_LEVEL,
    TS_URL_INFO_CACHE_PATH_BYTES,
    TS_URL_INFO_TLD,
    TS_URL_INFO_USER,
    TS_URL_INFO_PASSWORD,
    TS_URL_INFO_HAS_USERINFO,

    TS_URL_INFO_LAST /* do not use */
} TS_Url_Info;

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_UrlInfoGet(TS_Handle ts_handle,
              TS_Url url,
              TS_Url_Info info_type,
              void *info);


TS_API_PUBLIC int _cdecl
TS_UrlParse(TS_Handle ts_handle,
            const char *url,
            const char *host_header,
            TS_Url parsed_url);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */


/*--------------------------------------------------------------*\
 *
 *                     CUSTOM SEARCH KEYWORDS
 *
 \*--------------------------------------------------------------*/

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_CustomSearchKeywordsAdd(TS_Handle ts_handle,
                           const char *keyword,
                           TS_Categories categories);

TS_API_PUBLIC int _cdecl
TS_CustomSearchKeywordsPost(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_CustomSearchKeywordsRemove(TS_Handle ts_handle,
                              const char *keyword);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum TS_Custom_Search_Keywords_Reset_Type
{
    TS_CUSTOM_SEARCH_KEYWORDS_RESET_FIRST, /* Do not use */

    TS_CUSTOM_SEARCH_KEYWORDS_RESET_ACTIVE,
    TS_CUSTOM_SEARCH_KEYWORDS_RESET_PENDING,
    TS_CUSTOM_SEARCH_KEYWORDS_RESET_ALL,

    TS_CUSTOM_SEARCH_KEYWORDS_RESET_LAST   /* Do not use */
} TS_Custom_Search_Keywords_Reset_Type;

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_CustomSearchKeywordsReset(TS_Handle ts_handle,
                             TS_Custom_Search_Keywords_Reset_Type type);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */


/*--------------------------------------------------------------*\
 *
 *                     CUSTOM SITES
 *
 \*--------------------------------------------------------------*/

#define TS_CUSTOM_SITES_MATCH_PORT     0x00000001
#define TS_CUSTOM_SITES_MATCH_PROTOCOL 0x00000002

typedef enum
{
    TS_CUSTOM_SITES_TYPE_FIRST,  /* Do not use */

    TS_CUSTOM_SITES_TYPE_NONE,
    TS_CUSTOM_SITES_TYPE_SEARCH_PHRASE,

    TS_CUSTOM_SITES_TYPE_LAST   /* Do not use */
} TS_Custom_Sites_Data;


typedef enum TS_Custom_Sites_State
{
    TS_CUSTOM_SITES_STATE_FIRST, /* Do not use */

    TS_CUSTOM_SITES_STATE_INITIAL,
    TS_CUSTOM_SITES_STATE_FINAL,

    TS_CUSTOM_SITES_STATE_LAST   /* Do not use */
} TS_Custom_Sites_State;


#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_CustomSitesAdd(TS_Handle ts_handle,
                  TS_Url url,
                  int url_match,
                  TS_Categories categories,
                  TS_Custom_Sites_State state,
                  TS_Custom_Sites_Data data_type,
                  void *data);

TS_API_PUBLIC int _cdecl
TS_CustomSitesPost(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_CustomSitesRemove(TS_Handle ts_handle,
                     TS_Url url,
                     int match);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum TS_Custom_Sites_Reset_Type
{
    TS_CUSTOM_SITES_RESET_FIRST, /* Do not use */

    TS_CUSTOM_SITES_RESET_ACTIVE,
    TS_CUSTOM_SITES_RESET_PENDING,
    TS_CUSTOM_SITES_RESET_ALL,

    TS_CUSTOM_SITES_RESET_LAST   /* Do not use */
} TS_Custom_Sites_Reset_Type;



#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_CustomSitesReset(TS_Handle ts_handle,
                    TS_Custom_Sites_Reset_Type type);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */


/*--------------------------------------------------------------*\
 *
 *                       Network Cache
 *
 \*--------------------------------------------------------------*/

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int
TS_TSCacheClear(TS_Handle ts_handle);

TS_API_PUBLIC int
TS_TSCacheConfigure(unsigned int timeout,
                     unsigned int retry,
                     int max_entries,
                     char *file);

TS_API_PUBLIC int
TS_TSCacheLoad(TS_Handle ts_handle);

TS_API_PUBLIC int
TS_TSCacheStore(TS_Handle ts_handle);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                       LOGGING
 *
 \*--------------------------------------------------------------*/
typedef enum
{
    TS_LOG_LEVEL_FIRST = 0,

    TS_LOG_LEVEL_NONE,      /* Turns all logging off */
    TS_LOG_LEVEL_FATAL,     /* Bad error, going away or crashing */
    TS_LOG_LEVEL_ERROR,     /* Bad error, will try to recover */
    TS_LOG_LEVEL_WARNING,   /* Shouldn't have happened, but can recover */
    TS_LOG_LEVEL_INFO,      /* Something interesting (and normal) occurred */
    TS_LOG_LEVEL_DEBUG1,    /* DEBUG1 through DEBUG9 and also DEBUG */
    TS_LOG_LEVEL_DEBUG2,    /* represent the different debugging levels */
    TS_LOG_LEVEL_DEBUG3,    /* that an application may use.  Typically, */
    TS_LOG_LEVEL_DEBUG4,    /* the larger the number, the more logging */
    TS_LOG_LEVEL_DEBUG5,    /* you will see. */
    TS_LOG_LEVEL_DEBUG6,
    TS_LOG_LEVEL_DEBUG7,
    TS_LOG_LEVEL_DEBUG8,
    TS_LOG_LEVEL_DEBUG9,
    TS_LOG_LEVEL_DEBUG,     /* All debugging except for trace logs */
    TS_LOG_LEVEL_TRACE,     /* All debugging as well as trace logs */
    /* which mark the entrance and exit of */
    /* functions. */
    TS_LOG_LEVEL_ALL,       /* All log messages */

    TS_LOG_LEVEL_LAST
} TS_Log_Level;


typedef int TS_Log_Area;

#define TS_LOG_AREA_FIRST             TS_LOG_AREA_CUSTOM_SITES
#define TS_LOG_AREA_CUSTOM_SITES      0x00000001
#define TS_LOG_AREA_CUSTOM_KEYWORDS   0x00000002
#define TS_LOG_AREA_CUSTOM_PATTERNS   0x00000004
#define TS_LOG_AREA_DATABASE_DOWNLOAD 0x00000008
#define TS_LOG_AREA_DATABASE_LOAD     0x00000010
#define TS_LOG_AREA_DATABASE_SEARCH   0x00000020
#define TS_LOG_AREA_LOG               0x00000040
#define TS_LOG_AREA_URL               0x00000080
#define TS_LOG_AREA_CATEGORIES        0x00000100
#define TS_LOG_AREA_HANDLE            0x00000200
#define TS_LOG_AREA_ATTRIBUTES        0x00000400
#define TS_LOG_AREA_DNS               0x00000800
#define TS_LOG_AREA_NETWORK           0x00001000
#define TS_LOG_AREA_DOWNLOADTHREAD    0x00002000
#define TS_LOG_AREA_TRUSTEDSOURCE     0x00004000
#define TS_LOG_AREA_RATE_MESSAGE      0x00008000
#define TS_LOG_AREA_RATE_CONNECTION   0x00010000
#define TS_LOG_AREA_RATE_IPSPAM       0x00020000
#define TS_LOG_AREA_LAST              TS_LOG_AREA_RATE_IPSPAM

#define TS_LOG_AREA_MASK              (2 * TS_LOG_AREA_LAST - 1)

#define TS_LOG_AREA_ALL               TS_LOG_AREA_MASK

typedef void (*TS_Log_Func)(TS_Log_Level level,
                            TS_Log_Area area,
                            const char *message);

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_LogFunctionSet(TS_Handle ts_handle,
                  TS_Log_Func log_func);

TS_API_PUBLIC int _cdecl
TS_LogLevelSet(TS_Handle ts_handle,
               TS_Log_Level level,
               TS_Log_Area areas);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                       Web Database
 *
 \*--------------------------------------------------------------*/
typedef void *(*TS_Database_Alloc_Func)(int bytes);
typedef void (*TS_Database_Free_Func)(void *ptr);


#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DatabaseAccessFunctionsSet(TS_Handle ts_handle,
                              TS_Database_Alloc_Func database_alloc_func,
                              TS_Database_Free_Func database_free_func);

/*
 * Set security parameters for the database.
 *
 * Arguments:
 *
 * require_signed_downloads - if TS_ENABLE, will require signed databases when
 *     downloading. If the database downloaded is not signed, download will
 *     fail and the existing database will be used, if any. If TS_DISABLE, will
 *     download either signed or unsigned databases. The default is enabled
 *     (require signed downloads) and may also be specified as TS_USE_DEFAULT.
 *
 * require_signed_database - if TS_ENABLE, will only run with a signed database
 *     present. Attempting to load an unsigned database will fail. If
 *     TS_DISABLE, will load either signed or unsigned. The default is disabled
 *     (don't require signed databases) and may be specified as TS_USE_DEFAULT.
 *
 * max_database_age - the maximum age in days of a database. Applies only to
 *     downloads, not to loads. Older databases will refuse to download. May
 *     specify TS_USE_DEFAULT for the default of unlimited, which will download
 *     any database regardless of age.
 *
 * max_database_size - the maximum size in KB of the database. Applies only to
 *     downloads, not to loads. Larger databases will refuse to download. May
 *     specify TS_USE_DEFAULT for the default of unlimited, which will download
 *     any database no matter how big.
 *
 * It is an invalid combination to not require signed downloads, yet require
 * a signed database.
 *
 * Returns TS_OK on success, or TS_ERROR on error.
 */
TS_API_PUBLIC int
TS_DatabaseSetSecurity(TS_Handle    ts_handle,
                       int          require_signed_downloads,
                       int          require_signed_database,
                       unsigned int max_database_age,
                       unsigned int max_database_size);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum
{
    TS_DATABASE_DOWNLOAD_MODE_FIRST, /* do not use */

    TS_DATABASE_DOWNLOAD_MODE_FULL,
    TS_DATABASE_DOWNLOAD_MODE_INCR,
    TS_DATABASE_DOWNLOAD_MODE_INCR_LOW_CPU,

    TS_DATABASE_DOWNLOAD_MODE_LAST  /* do not use */

} TS_Database_Download_Mode;

#define TS_DATABASE_DOWNLOAD_COMPLETE   1
#define TS_DATABASE_DOWNLOAD_PARTIAL    2
#define TS_DATABASE_DOWNLOAD_NOT_NEEDED 3

#define TS_DOWNLOAD_STATUS_NOT_DOWNLOADING      0
#define TS_DOWNLOAD_STATUS_INITIALIZING         1
#define TS_DOWNLOAD_STATUS_DOWNLOADING          2
#define TS_DOWNLOAD_STATUS_MERGING              3
#define TS_DOWNLOAD_STATUS_COMPLETE             4
#define TS_DOWNLOAD_STATUS_FAILED               5
#define TS_DOWNLOAD_STATUS_CANCELLED            6

#define TS_DATABASE_DOWNLOAD_DEFAULT_HOST   "list.smartfilter.com"
#define TS_DATABASE_DOWNLOAD_DEFAULT_PORT   80
#define TS_DATABASE_DOWNLOAD_DEFAULT_PATH   "cgi-bin/updatelist"

#define TS_PROXY_NAME_LEN_MAX

typedef enum TS_Database_Download_Error
{
    TS_DATABASE_DOWNLOAD_ERROR_FIRST, /* do not use */

    TS_DATABASE_DOWNLOAD_ERROR_INTERNAL,
    TS_DATABASE_DOWNLOAD_ERROR_RESPONSE_INVALID,
    TS_DATABASE_DOWNLOAD_ERROR_FILE_PERMISSIONS,
    TS_DATABASE_DOWNLOAD_ERROR_CONNECT,
    TS_DATABASE_DOWNLOAD_ERROR_HOSTNAME_INVALID,
    TS_DATABASE_DOWNLOAD_ERROR_HTTP_STATUS,
    TS_DATABASE_DOWNLOAD_ERROR_CANCELLED,
    TS_DATABASE_DOWNLOAD_ERROR_DISABLED,
    TS_DATABASE_DOWNLOAD_ERROR_CONTENT_TOO_LARGE,
    TS_DATABASE_DOWNLOAD_ERROR_CONTENT_TOO_OLD,

    TS_DATABASE_DOWNLOAD_ERROR_LAST   /* do not use */
} TS_Database_Download_Error;


typedef enum TS_Database_Download_Database_Type
{
    TS_DATABASE_DOWNLOAD_DATABASE_TYPE_FIRST, /* do not use */

    TS_DATABASE_DOWNLOAD_DATABASE_TYPE_STANDARD,
    TS_DATABASE_DOWNLOAD_DATABASE_TYPE_XL,
    TS_DATABASE_DOWNLOAD_DATABASE_TYPE_TS,

    TS_DATABASE_DOWNLOAD_DATABASE_TYPE_LAST   /* do not use */
} TS_Database_Download_Database_Type;


#define TS_PRODUCT_TYPE_LEN    80
#define TS_PRODUCT_VERSION_LEN 20

typedef struct TS_Database_Download_Func_Info
{
    TS_Handle ts_handle;
    char dest_host[80]; /* must fit TS_DATABASE_DOWNLOAD_DEFAULT_HOST */
    unsigned short dest_port;
    char proxy_host[80];
    unsigned short proxy_port;
    char proxy_username[80];
    char proxy_password[80];
    char path[1024]; /* must fit TS_DATABASE_DOWNLOAD_DEFAULT_PATH */
    char serial_number[80];
    char perm_serial_number[80];
    TS_Database_Download_Database_Type database_type;
    char product_type[TS_PRODUCT_TYPE_LEN];
    char product_version[TS_PRODUCT_VERSION_LEN];
    char os[20];
    char *distributes_to;  /* For internal use only */
    int user_count;
    char expiration[80];
    TS_Database_Download_Error status;
    int http_status;
    char message[1024];
    TS_Log_Level message_level;
    int message_code;
    void *cb; /* for callback implementation in Java */
} TS_Database_Download_Func_Info;


typedef int (* TS_Download_Complete_Callback_Func)(int retcode,
                                                   int status,
                                                   void *userdata);

typedef int (* TS_Download_Progress_Update_Callback_Func)(int phase,
                                                          int step_current,
                                                          int step_max,
                                                          int substep_current,
                                                          int substep_max,
                                                          void *userdata);

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DownloadThreadStart(TS_Handle ts_handle,
                       const char *local_filename,
                       TS_Download_Complete_Callback_Func callback_func,
                       void *data);

TS_API_PUBLIC int _cdecl
TS_DownloadThreadStop(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_DownloadCancel(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_DownloadStart(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_DownloadCancelled(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_DownloadProgressUpdate(TS_Handle ts_handle,
                          unsigned int current,
                          unsigned int max);

TS_API_PUBLIC int _cdecl
TS_DownloadProgressGet(TS_Handle ts_handle,
                       int *current_primary,
                       int *max_primary,
                       int *current_secondary,
                       int *max_secondary,
                       int *phase);

TS_API_PUBLIC int _cdecl
TS_DownloadProgressReset(TS_Handle ts_handle);

TS_API_PUBLIC int _cdecl
TS_DownloadProgressCallbackRegister(TS_Handle ts_handle,
                                    TS_Download_Progress_Update_Callback_Func func,
                                    unsigned int interval,
                                    void *data);

TS_API_PUBLIC int _cdecl
TS_DatabaseDownload(TS_Handle ts_handle,
                    const char *local_filename,
                    TS_Database_Download_Mode download_mode,
                    int *download_status,
                    void *data);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

#define TS_FILE_TYPE_SUBSCRIPTION -4
#define TS_FILE_TYPE_INFO         -3
#define TS_FILE_TYPE_FULL         -2
#define TS_FILE_TYPE_CURRENT      -1

typedef int (*TS_Database_Download_Func)(const char *local_file,
                                         int file_type,
                                         void *data);

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DatabaseDownloadFunctionSet(TS_Handle ts_handle,
                               TS_Database_Download_Func download_func);


TS_API_PUBLIC int _cdecl
TS_DatabaseSubscriptionInfoDownload(TS_Handle ts_handle,
                                    TS_Database_Download_Func_Info *data);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum
{
    TS_DATABASE_ACCESS_FIRST, /* do not use */

    TS_DATABASE_ACCESS_DISK,
    TS_DATABASE_ACCESS_MEMORY,
    TS_DATABASE_ACCESS_EXTERNAL,
    TS_DATABASE_ACCESS_BUFFER,

    TS_DATABASE_ACCESS_LAST   /* do not use */
} TS_Database_Access;

#define TS_CAT_SET_LATEST ((unsigned int)-1)
#define TS_CAT_SET_ALL    ((unsigned int)-2)
#define TS_CAT_SET_LOADED ((unsigned int)-3)

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DatabaseLoad(TS_Handle ts_handle,
                const char *filename,
                TS_Database_Access access,
                unsigned int category_set_version);

TS_API_PUBLIC int _cdecl
TS_DatabaseLoadFromMemory(TS_Handle    ts_handle,
                          const char  *filename, /* Only used for downloads */
                          const char  *mem_ptr,
                          unsigned int mem_len,
                          unsigned int category_set_version);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

#if !defined(WIN32)
#if TS_FEATURE_FUNCTION_DECLARATIONS

/* Not implemented on Windows to allow ABI backwards compatibility  -- enable when possible */
TS_API_PUBLIC int _cdecl
TS_DatabaseLoadFromMemoryExtended(TS_Handle    ts_handle,
                                  const char  *filename,
                                  const char  *mem_ptr,
                                  unsigned int mem_len,
                                  unsigned int category_set_version,
                                  int crc_check);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */
#endif

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DatabaseLoadFromHandle(TS_Handle ts_handle_src,
                          TS_Handle ts_handle_dst);

TS_API_PUBLIC int _cdecl
TS_DatabaseReload(TS_Handle ts_handle,
                  const char *filename);

TS_API_PUBLIC int _cdecl
TS_DatabaseReloadFromMemory(TS_Handle    ts_handle,
                            const char  *mem_ptr,
                            unsigned int mem_len);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

#if !defined(WIN32)
#if TS_FEATURE_FUNCTION_DECLARATIONS

/* Not implemented on Windows to allow ABI backwards compatibility  -- enable when possible */
TS_API_PUBLIC int _cdecl
TS_DatabaseReloadFromMemoryExtended(TS_Handle    ts_handle,
                                    const char  *mem_ptr,
                                    unsigned int mem_len,
                                    int crc_check);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */
#endif

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DatabaseUnload(TS_Handle ts_handle);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */


/*--------------------------------------------------------------*\
 *
 *                         PATTERNS
 *
 \*--------------------------------------------------------------*/
typedef enum
{
    TS_CATEGORIES_ACTION_FIRST, /* do not use */

    TS_CATEGORIES_ACTION_OVERRIDE,
    TS_CATEGORIES_ACTION_AUGMENT,

    TS_CATEGORIES_ACTION_LAST /* do not use */
} TS_Patterns_Categories_Action;



#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_PatternsAdd(TS_Handle ts_handle,
               TS_Categories input_categories,
               const char *protocol,
               const char *domain_pattern,
               const char *path_pattern,
               TS_Categories output_categories,
               TS_Patterns_Categories_Action categories_action);

TS_API_PUBLIC int _cdecl
TS_PatternsPost(TS_Handle ts_handle);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

typedef enum TS_Patterns_Reset_Type
{
    TS_PATTERNS_RESET_FIRST, /* Do not use */

    TS_PATTERNS_RESET_ACTIVE,
    TS_PATTERNS_RESET_PENDING,
    TS_PATTERNS_RESET_ALL,

    TS_PATTERNS_RESET_LAST   /* Do not use */
} TS_Patterns_Reset_Type;



#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_PatternsReset(TS_Handle ts_handle,
                 TS_Patterns_Reset_Type type);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                         EXTERNAL DNS
 *
 \*--------------------------------------------------------------*/

#define TS_FORWARD    1
#define TS_REVERSE    2

#define TS_IPV6       1
#define TS_IPV4       2
#define TS_IPVANY     3

typedef enum
{
    TS_DNSQUERY_TYPE_NONE = 0, /* Do not use */
    TS_DNSQUERY_TYPE_DOMAIN,
    TS_DNSQUERY_TYPE_IPV4,
    TS_DNSQUERY_TYPE_IPV6,
    TS_DNSQUERY_TYPE_IPANY
} TS_DNSQueryType;

#define TS_DNS_MAX_IP 16

typedef int (* ts_dns_func)(char *target,
                            int ip_type,
                            int direction,
                            char *name,
                            int namelen,
                            struct sockaddr_storage *addr,
                            int maxaddrs,
                            int *addrcount,
                            void *userdata);

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int _cdecl
TS_DNSEnable(TS_Handle ts_handle, int value);


TS_API_PUBLIC int _cdecl
TS_DNSRegisterExternal(TS_Handle ts_handle,
                       ts_dns_func func,
                       void *userdata);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                         Networking
 *
 \*--------------------------------------------------------------*/

/*
 * Return
 */
#define TS_NET_DNS_FAILURE              2
#define TS_NET_OK                       1
#define TS_NET_ERROR                    -1
#define TS_NET_NOT_READY                -2

#define TS_NETLOOKUP_SERVER_DEFAULT     NULL
#define TS_NETLOOKUP_PORT_DEFAULT       -1

/*
 * Events
 */
typedef enum
{
    TS_NET_EVENT_FIRST = 1, /* do not use */

    TS_NET_EVENT_READ,
    TS_NET_EVENT_READ_DATA,
    TS_NET_EVENT_READ_READY,
    TS_NET_EVENT_WRITE_READY,
    TS_NET_EVENT_ERROR,
    TS_NET_EVENT_CLOSE,

    TS_NET_EVENT_LAST /* do not use */
} TS_NetEvent;



typedef int (* TS_NetLookupExternalConnectFunc)(int conn_id,
                                                void *callback_data,
                                                void **conn_data);

typedef int (* TS_NetLookupExternalSendFunc)(int conn_id,
                                             const char *data,
                                             int data_len,
                                             void *callback_data,
                                             void *conn_data);

typedef int (*TS_NetLookupExternalReceiveFunc)(int conn_id,
                                               char *data,
                                               int max_data_size,
                                               void *callback_data,
                                               void *conn_data);

typedef int (*TS_NetLookupExternalCloseFunc)(int conn_id,
                                             void *callback_data,
                                             void *conn_data);

#if TS_FEATURE_FUNCTION_DECLARATIONS

/* [ts_connection_sharing] allowed */
TS_API_PUBLIC int _cdecl
TS_NetLookupConfigureInternal(TS_Handle ts_handle,
                              const char *device_id,
                              const char *host,
                              short port,
                              const char *cert,
                              int cert_len,
                              const char *privkey,
                              int privkey_len,
                              const char *cacert,
                              int cacert_len);

/* [ts_connection_sharing] allowed */
TS_API_PUBLIC int _cdecl
TS_NetLookupConfigureHTTPProxy(TS_Handle       ts_handle,
                               const char     *proxy_host,
                               unsigned short  proxy_port,
                               const char     *proxy_username,
                               const char     *proxy_password);

/* [ts_connection_sharing] allowed */
/* FIXME: does 'disabling' lookups make sense? */
TS_API_PUBLIC int _cdecl
TS_NetLookupEnable(TS_Handle ts_handle,
                   int value);

/* [ts_connection_sharing] allowed (possibly implemented in the future) */
TS_API_PUBLIC int _cdecl
TS_NetLookupExternalFunctionsSet(TS_Handle ts_handle,
                                 TS_NetLookupExternalConnectFunc conn_func,
                                 TS_NetLookupExternalSendFunc send_func,
                                 TS_NetLookupExternalReceiveFunc recv_func,
                                 TS_NetLookupExternalCloseFunc close_func,
                                 void *callback_data);

TS_API_PUBLIC int _cdecl
TS_NetLookupConfigureSettings(TS_Handle ts_handle,
                              int connections,
#if defined(TS_USE_LONG_LONG_TIMEOUT)
                              long long timeout,
#else
                              long timeout,
#endif
                              int attempts);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/* possible values for TS_HTTPProxyCallbackSettings::hpcb_action {{{ */
#define TS_HTTPPROXY_CALLBACK_ACTION_USE_DEFAULT                0       /* use default proxy settings, if any; (default) */
#define TS_HTTPPROXY_CALLBACK_ACTION_USE_ALTERNATE              1       /* use alternate proxy as specified in TS_HTTPProxyCallbackSettings; */
/* }}} */

typedef struct
{
	unsigned int    structure_size;     /* initialise to sizeof TS_NetLookupHTTPProxyAuthData ) */
	char*           proxy_host;
	unsigned short  proxy_port;
	char*           domain_name;         /*for NTLM and Kerberos authentication*/
	char*           proxy_username;      /*for NTLM and Kerberos it is a domain user name */
	char*           proxy_password;
	char*           kerberos_principal;    /* Kerberos(optional): principal name for the client (username/instance@REALM) where the REALM could be a windows domain, default: user/domain_name@domain_name’*/
	char*           kerberos_targetname;   /* Kerberos(optional): target name for the Proxy, default value: “http/proxyhost@domain_name”*/
	unsigned int    proxy_connection_timeout;   /* Proxy connection timeout (optional): default value: 1000ms */
} TS_NetLookupHTTPProxyAuthData;

/* TS_HTTPProxyCallbackSettings */
typedef struct
{
    unsigned int                        hpcb_version;               /* version of TS_HTTPProxyCallbackSettings structure user
                                                                     * should make sure that this version is equal to or greater
                                                                     * than the version of field it is going to access */
    /* v1 -> */
    const char *                        hpcb_host_name;             /* host name null terminated string */
    const struct sockaddr_storage *     hpcb_host_ip;               /* host ip */
    void *                              hpcb_user_data;             /* user data (TS_NetLookupSettings::http_proxy_callback_user_data) */

    char *                              hpcb_proxy_host;            /* proxy host name to use */
    size_t                              hpcb_proxy_host_size;       /* size of proxy host buffer in bytes */
    unsigned short                      hpcb_proxy_port;            /* proxy port to use */
    char *                              hpcb_proxy_username;        /* proxy username */
    size_t                              hpcb_proxy_username_size;   /* proxy username size */
    char *                              hpcb_proxy_password;        /* proxy password */
    size_t                              hpcb_proxy_password_size;   /* proxy password size */
    unsigned int                        hpcb_action;                /* action: to indiate action taken by callback
                                                                     * see TS_HTTPPROXY_CALLBACK_ACTION_* for possible values */
    /* <- v1 */
} TS_HTTPProxyCallbackSettings;

typedef struct
{
	TS_HTTPProxyCallbackSettings    settings;                       /*legacy callback data*/
	unsigned int                    structure_size;                 /* initialize to sizeof TS_HTTPProxyCallbackSettingsEx */
	unsigned int                    http_proxy_flags;               /*same as in TS_NetLookupSettings */
	char*                           hpcbx_domain_name;
    size_t                          hpcbx_domain_name_size;
	char*                           hpcbx_kerberos_principal;
	size_t                          hpcbx_kerberos_principal_size;
	char*                           hpcbx_kerberos_targetname;
	size_t                          hpcbx_kerberos_target_size;
	unsigned int					hpcbx_proxy_connection_timeout;
} TS_HTTPProxyCallbackSettingsEx;

/* http proxy callback definition
 * returns: non-zero on success, 0 on failure
 * on failure, connection setup process is aborted
 */
typedef int
(* TS_NetLookupHTTPProxyCallbackFunc )(
        TS_Handle ts_handle,                                        /* [IN]     related ts handle */
        TS_HTTPProxyCallbackSettings * settings_ptr                 /* [IN/OUT] proxy settings structure */
    );

typedef int
(*TS_NetLookupHTTPProxyCallbackFuncEx)(
TS_Handle ts_handle,                                          /* [IN]     related ts handle */
TS_HTTPProxyCallbackSettingsEx * settings_ptr                 /* [IN/OUT] proxy settings structure */
);

/*
 * Mask and structure definition for TS_NetLookupConfigureSettingsExtended below
 */

/* member masks */
/* v1 -> */
#define TS_NETLOOKUPCONFIGEXT_CONNECTIONS                                       0x0001U /* member: connections       */
#define TS_NETLOOKUPCONFIGEXT_REQUEST_TIMEOUT                                   0x0002U /* member: request_timeout   */
#define TS_NETLOOKUPCONFIGEXT_REQUEST_ATTEMPTS                                  0x0004U /* member: request_attempts  */
#define TS_NETLOOKUPCONFIGEXT_SSL_WRITE_TIMEOUT                                 0x0008U /* member: ssl_write_timeout */
/* <- v1 */
/* v2 -> */
#define TS_NETLOOKUPCONFIGEXT_IPC_MODE                                          0x0010U /* member: ipc_mode */
#define TS_NETLOOKUPCONFIGEXT_IPC_OSNAME                                        0x0020U /* member: ipc_osname */
#define TS_NETLOOKUPCONFIGEXT_IPC_SERVER_MAXCLIENTS                             0x0040U /* member: ipc_server_maxclients */
#define TS_NETLOOKUPCONFIGEXT_IPC_CLIENT_RESOURCE_PROBING_INTERVAL_MIN          0x0080U /* member: ipc_client_resource_probing_interval_min */

#if defined(WIN32) || defined(WIN64)

#define TS_NETLOOKUPCONFIGEXT_IPC_SERVER_WINDOWS_RESOURCE_CREATION_SECATTR      0x0100U /* member: ipc_server_windows_resource_creation_secattr */

#else /* non-windows */
#endif /* windows/non-windows */

#define TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_FUNC                           0x0200U /* member: http_proxy_callback_func */
#define TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_ALLOC_STRING_SIZE              0x0400U /* member: http_proxy_callback_alloc_string_size */
#define TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_USER_DATA                      0x0800U /* member: http_proxy_callback_user_data */
#define TS_NETLOOKUPCONFIGEXT_HTTPPROXY_FLAGS                                   0x1000U /* member: http_proxy_flags */

#define TS_NETLOOKUP_IPC_MODE_DISABLED                                          0       /* no IPC functionality */
#define TS_NETLOOKUP_IPC_MODE_CLIENT                                            1       /* enable client-side functionality (default) */
#define TS_NETLOOKUP_IPC_MODE_SERVER                                            2       /* act as an IPC server */

/* Values for ipc_server_maxclients */
#define TS_NETLOOKUP_IPC_SERVER_MAXCLIENTS_DEFAULT                              0       /* Use default for max clients */

/* Values for http_proxy_flags */
/* Remember to update LOCAL_HTTPPROXY_FLAGS_ALL in ts_network.c */
#define TS_HTTPPROXY_FLAGS_USETSHOSTNAME                                        0x0001U /* Use hostname instead of numeric IP */

/* Flags for setting proxy authentication sequence */
/* If the flag is set the code will try only Kerberos authentication */
#define TS_HTTPPROXY_FLAGS_KERB_AUTH                                            0x1000U

/* If the flag is set the code will try Kerberos first and if it fails it will try Basic next */
#define TS_HTTPPROXY_FLAGS_KERB_BASIC_AUTH										0x1200U

/* If the flag is set the code will try Kerberos first and if it fails it will try NTLM next */
#define TS_HTTPPROXY_FLAGS_KERB_NTLM_AUTH										0x1800U

/* If the flag is set the code will try Kerberos first and if it fails it will try Basic next and if it fails it will try NTLM as last option */
#define TS_HTTPPROXY_FLAGS_KERB_BASIC_NTLM_AUTH									0x1e00U

/* If the flag is set the code will try Kerberos first and if it fails it will try NTLM next and if it fails it will try Basic as last option */
#define TS_HTTPPROXY_FLAGS_KERB_NTLM_BASIC_AUTH									0x1b00U

/* If the flag is set the code will try only NTLM authentication */
#define TS_HTTPPROXY_FLAGS_NTLM_AUTH											0x0400U

/* If the flag is set the code will try NTLM first and if it fails it will try Basic next */
#define TS_HTTPPROXY_FLAGS_NTLM_BASIC_AUTH										0x0600U

/* If the flag is set the code will try NTLM first and if it fails it will try Kerberos next */
#define TS_HTTPPROXY_FLAGS_NTLM_KERB_AUTH										0x2400U

/* If the flag is set the code will try NTLM first and if it fails it will try Basic next and if it fails it will try Kerberos as last option */
#define TS_HTTPPROXY_FLAGS_NTLM_BASIC_KERB_AUTH									0x3600U

/* If the flag is set the code will try NTLM first and if it fails it will try Kerberos next and if it fails it will try Basic as last option */
#define TS_HTTPPROXY_FLAGS_NTLM_KERB_BASIC_AUTH									0x2700U

/* If the flag is set the code will try only Basic authentication */
#define TS_HTTPPROXY_FLAGS_BASIC_AUTH											0x0100U

/* If the flag is set the code will try Basic first and if it fails it will try NTLM next */
#define TS_HTTPPROXY_FLAGS_BASIC_NTLM_AUTH										0x0900U

/* If the flag is set the code will try Basic first and if it fails it will try Kerberos next */
#define	TS_HTTPPROXY_FLAGS_BASIC_KERB_AUTH										0x2100U

/* If the flag is set the code will try Basic first and if it fails it will try NTLM next and if it fails it will try Kerberos as last option */
#define TS_HTTPPROXY_FLAGS_BASIC_NTLM_KERB_AUTH									0x3900U

/* If the flag is set the code will try Basic first and if it fails it will try Kerberos next and if it fails it will try NTLM as last option */
#define TS_HTTPPROXY_FLAGS_BASIC_KERB_NTLM_AUTH									0x2d00U

/* If the flag is set the code will skip the feature of checking the proxy supported authentications before trying any authentication */
#define TS_HTTPPROXY_FLAGS_SKIP_PROXY_SUPPORTED_AUTH_CHECK						0x8000U

/* The default sequence for 2.5.0.1 would be Kerberos-->NTLM-->Basic */
#define TS_HTTPPROXY_FLAGS_DEFAULT_AUTH    TS_HTTPPROXY_FLAGS_KERB_NTLM_BASIC_AUTH       

#define LOCAL_HTTPPROXY_FLAGS_ALL TS_HTTPPROXY_FLAGS_USETSHOSTNAME | TS_HTTPPROXY_FLAGS_BASIC_NTLM_KERB_AUTH | TS_HTTPPROXY_FLAGS_BASIC_KERB_NTLM_AUTH | TS_HTTPPROXY_FLAGS_KERB_NTLM_BASIC_AUTH | TS_HTTPPROXY_FLAGS_SKIP_PROXY_SUPPORTED_AUTH_CHECK
/* <- v2 */

/* v3 -> */
#if defined(WIN32) || defined(WIN64)

#define TS_NETLOOKUPCONFIGEXT_IPC_CFG_WINREG_KEY_HANDLE                         0x2000U /* member: ipc_cfg_winreg_key_handle */
#define TS_NETLOOKUPCONFIGEXT_IPC_CFG_WINREG_SUBKEY                             0x4000U /* member: ipc_cfg_winreg_subkey */
#define TS_NETLOOKUPCONFIGEXT_IPC_CFG_WINREG_SECATTR                            0x8000U /* member: ipc_cfg_winreg_secattr */

#else /* non-windows */
#endif /* windows/non-windows */

#define TS_NETLOOKUPCONFIGEXT_FLAGS                                             0x00010000U /* members: flags */

/* possible values of flags {{{ */
typedef unsigned int ts_netlookup_flags_t;
#define TS_NETLOOKUP_FLAGS_value_cast( val )                                    ( ( ts_netlookup_flags_t ) val )
/* backing off flags GROUP {{{ */
#define TS_NETLOOKUP_FLAGS_BACKOFF_OPTION_DEFAULT                               TS_NETLOOKUP_FLAGS_value_cast( 0x00000001 )
#define TS_NETLOOKUP_FLAGS_BACKOFF_OPTION_FAIL_REQ_IMMEDIATELY                  TS_NETLOOKUP_FLAGS_value_cast( 0x00000002 )
/* }}} backing off flags GROUP */

/* Connection directives flags GROUP {{{ */
#define TS_NETLOOKUP_FLAGS_USE_DIRECT_CONNECTION_IF_PROXY_UNREACHABLE           TS_NETLOOKUP_FLAGS_value_cast( 0x00000004 )
/* }}} Connection directives flags GROUP */

/* }}} possible values of flags */
/* <- v3 */

/* v4 -> */

#if defined( WIN32 ) || defined( WIN64 )
#define TS_NETLOOKUPCONFIGEXT_HTTPPROXY_NTLM_USER_TOKEN                         0x00020000U /* member: http_proxy_ntlm_user_token */
#endif /* windows */

/* <- v4 */

/* v5 -> */
#define TS_NETLOOKUPCONFIGEXT_PROXY_AUTH_DATA                                   0x00080000U /* member: proxy_auth_data */
#define TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_FUNC_EX                        0x00100000U /* member: http_proxy_callback_func () */
/* <- v5 */

/* TS_NetLookupSettings V4 */
typedef struct
{
    /* v1 -> */
    unsigned int    structure_size;     /* initialise to sizeof( TS_NetLookupSettings ) */
    unsigned int    members_mask;       /* set value by "or"-ing masks describing structure fields  */
                                        /* (TS_NETLOOKUPCONFIGEXT_* symbols) set by caller */

    /*
     * fields that can be selected by specifying the right mask (see members_mask)
     */
    unsigned int    connections;        /* max connections to create. mask  TS_NETLOOKUPCONFIGEXT_CONNECTIONS       */
    unsigned int    request_timeout;    /* roundtrip timeout (us). mask:    TS_NETLOOKUPCONFIGEXT_REQUEST_TIMEOUT   */
    unsigned int    request_attempts;   /* retries per request. mask:       TS_NETLOOKUPCONFIGEXT_REQUEST_ATTEMPTS  */
    unsigned int    ssl_write_timeout;  /* ssl write timeout (ms). mask:    TS_NETLOOKUPCONFIGEXT_SSL_WRITE_TIMEOUT */
    /* <- v1 */

    /* v2 -> */
    unsigned int    ipc_mode;           /* IPC 'mode': values: TS_NETLOOKUP_IPC_MODE_* */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_IPC_MODE */
    const char *    ipc_osname;         /* basename for the IPC resource (currently: a named pipe) */
                                        /*  note: no pathname separators are allowed */
                                        /*  Currently TS_IPC_TRANSPORT_INTERNAL_DEFAULT_PIPE_BASENAME is always used. */
                                        /*  It is not currently user-configurable. */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_IPC_OSNAME */
    unsigned int    ipc_server_maxclients;
                                        /* server mode: maximum number of clients */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_IPC_SERVER_MAXCLIENTS */
    unsigned int    ipc_client_resource_probing_interval_min;
                                        /* client mode: minimum time interval to probe for an ipc server */
                                        /*  (in milliseconds) */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_IPC_CLIENT_RESOURCE_PROBING_INTERVAL_MIN */
#if defined(WIN32) || defined(WIN64)
    LPSECURITY_ATTRIBUTES   ipc_server_windows_resource_creation_secattr;
                                        /* server mode: (optional) security attributes to be used for IPC resource creation */
                                        /*  note: attributes (and its resources) are owned by caller */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_IPC_SERVER_WINDOWS_RESOURCE_CREATION_SECATTR */
#else /* non-windows */
#endif /* windows/non-windows */

#ifdef TS_SDK_2_5
	/*can be TS_NetLookupHTTPProxyCallbackFuncEx or TS_NetLookupHTTPProxyCallbackFunc*/
	void*                             http_proxy_callback_func;
#else
	/*old definition for backward compatibility*/
    TS_NetLookupHTTPProxyCallbackFunc http_proxy_callback_func;
#endif
	/* TS_NetLookupHTTPProxyCallbackFuncEx if TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_FUNC_EX is used*/
                                        /* proxy callback: (optional) called everytime a connection is about to be created
                                         * following fields are only used if http_proxy_callback_func field is set:
                                         *  -   http_proxy_callback_alloc_string_size
                                         *  -   http_proxy_callback_user_data
                                         */
                                        /* see: definition of TS_NetLookupHTTPProxyCallbackFunc for details */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_FUNC */

    size_t http_proxy_callback_alloc_string_size;
                                        /* output strings alloc size: (optional) size to allocate for each
                                         * of the following output string fields in the TS_HTTPProxyCallbackSettings
                                         * structure when passed in with 'http_proxy_callback_func':
                                         *  -   hpcb_proxy_host
                                         *  -   hpcb_proxy_username
                                         *  -   hpcb_proxy_password
                                         * if this is not defined then an appropriate default value will be used */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_ALLOC_STRING_SIZE */

    void * http_proxy_callback_user_data;
                                        /* pointer to user data for callback: (optional)
                                         * default value is NULL */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_HTTPPROXY_CALLBACK_USER_DATA */
    unsigned int http_proxy_flags;
                                        /* proxy flags: (optional) for future expansion, not used */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_HTTPPROXY_FLAGS */
                                        /* see TS_HTTPPROXY_FLAGS_* for possible values */
    /* <- v2 */

    /* v3 -> */
#if defined(WIN32) || defined(WIN64)

    HKEY    ipc_cfg_winreg_key_handle;
                                        /* handle to registry key: (optional) can be a predefined HIVE constant or handle
                                         * to any arbitrary key. handle is owned by the caller.
                                         * NOTE: if HKEY_CURRENT_USER is specified, RegOpenCurrentUser() will be used
                                         * to retrieve so that user impersonation works correctly.
                                         * Cannot be NULL
                                         * mask: TS_NETLOOKUPCONFIGEXT_IPC_CFG_WINREG_KEY_HANDLE
                                         */

    LPCTSTR  ipc_cfg_winreg_subkey;
                                        /* registry subkey: (optional) to used in combination with ipc_cfg_windows_reg_key_handle
                                         * to store/retrieve transport identifier in Windows Registry. subkey is owned by caller
                                         * Cannot be NULL or empty
                                         * mask: TS_NETLOOKUPCONFIGEXT_IPC_CFG_WINREG_SUBKEY
                                         */

    LPSECURITY_ATTRIBUTES   ipc_cfg_winreg_secattr;
                                        /* registry security attributes: (optional) security attributes to be used with.
                                         * Can be NULL to specify default.
                                         * attributes (and its resources) are owned by caller
                                         * mask: TS_NETLOOKUPCONFIGEXT_IPC_CFG_WINREG_SECATTR
                                         */

#else /* non-windows */
#endif /* windows/non-windows */

    ts_netlookup_flags_t    flags;      /* general flags: (optional) used to configure various options. caller
                                         * specifies one option from each group. if no option is specified for
                                         * a group then it remains unchanged.
                                         * for possible values see: TS_NETLOOKUP_FLAGS_*
                                         * mask: TS_NETLOOKUPCONFIGEXT_FLAGS
                                         */
    /* <- v3 */

    /* v4 -> */
#if defined( WIN32 ) || defined( WIN64 )
    HANDLE http_proxy_ntlm_user_token;
                                        /* ntlm user token: (optional) to be used duing NTLM authentication */
                                        /* note: owned by caller */
                                        /* mask: TS_NETLOOKUPCONFIGEXT_HTTPPROXY_NTLM_USER_TOKEN */
#endif /* windows */
    /* <- v4 */
	/* v5 -> */
	/* v5  is also used by REST! */
	//TS_NetLookupRestClient* rest_client_data;
	TS_NetLookupHTTPProxyAuthData* proxy_auth_data;
	/* <- v5 */
}   TS_NetLookupSettings_v5;

typedef TS_NetLookupSettings_v5 TS_NetLookupSettings;

#if TS_FEATURE_FUNCTION_DECLARATIONS

/* [ts_connection_sharing] allowed */
TS_API_PUBLIC int _cdecl
TS_NetLookupConfigureSettingsExtended(TS_Handle ts_handle,
                                      const TS_NetLookupSettings * config_data);

TS_API_PUBLIC int _cdecl
TS_NetLookupNotify(TS_Handle ts_handle,
                   int conn_id,
                   TS_NetEvent event,
                   char *message);

TS_API_PUBLIC int _cdecl
TS_NetLookupNotifyData(TS_Handle ts_handle,
                       int conn_id,
                       TS_NetEvent event,
                       const char *data,
                       int data_len);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                          Activation
 *
 \*--------------------------------------------------------------*/
#define TS_ACTIVATION_SERVER_DEFAULT NULL

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int
TS_ActivateTrustedSourceSetCA(TS_Handle ts_handle,
                              const char *filename,
                              int noverify);

TS_API_PUBLIC int
TS_ActivateTrustedSource(TS_Handle ts_handle,
                         const char *activation_server_url,
                         const char *product_type,
                         const char *product_version,
                         const char **formatted_serial,
                         const char **client_cert,
                         const char **client_key,
                         const char **ts_server_cert,
                         const char **errors);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                         Mail
 *
 \*--------------------------------------------------------------*/

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int
TS_RateMessageExtended(TS_Handle ts_handle,
                       const char *ehlo,
                       unsigned int ehlo_length,
                       struct sockaddr_storage source_ip,
                       const char *from,
                       unsigned int from_length,
                       const char *to,
                       unsigned int to_length,
                       const char *message,
                       unsigned int message_length,
                       const char *subject,
                       unsigned int subject_length,
                       const char *body,
                       unsigned int body_length,
                       int flags,
                       TS_Attributes attributes,
                       TS_RatingSession *session,
                       TS_RatingCompleteFunc callback,
                       void *callback_data);

TS_API_PUBLIC int
TS_RateMessage(TS_Handle ts_handle,
               const char *ehlo,
               unsigned int ehlo_length,
               struct sockaddr_storage source_ip,
               const char *from,
               unsigned int from_length,
               const char *to,
               unsigned int to_length,
               const char *message,
               unsigned int message_length,
               int flags,
               TS_Attributes attributes);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

/*--------------------------------------------------------------*\
 *
 *                         Connections
 *
 \*--------------------------------------------------------------*/

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int
TS_RateConnectionExtended(TS_Handle ts_handle,
                          struct sockaddr_storage src_ip,
                          unsigned short src_port,
                          struct sockaddr_storage dst_ip,
                          unsigned short dst_port,
                          unsigned char transport_proto,
                          int flags,
                          TS_Attributes attributes,
                          TS_RatingSession *session,
                          TS_RatingCompleteFunc callback,
                          void *callback_data);

TS_API_PUBLIC int
TS_RateConnection(TS_Handle ts_handle,
                  struct sockaddr_storage src_ip,
                  unsigned short src_port,
                  struct sockaddr_storage dst_ip,
                  unsigned short dst_port,
                  unsigned char transport_proto,
                  int flags,
                  TS_Attributes attributes);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

#define TS_TRANSPORT_TCP      1
#define TS_TRANSPORT_UDP      2
#define TS_TRANSPORT_ICMP     3
#define TS_TRANSPORT_UNKNOWN  10
/*--------------------------------------------------------------*\
 *
 *                         IP for SPAM
 *
 \*--------------------------------------------------------------*/

#if TS_FEATURE_FUNCTION_DECLARATIONS

TS_API_PUBLIC int
TS_RateIPForSpamExtended(TS_Handle ts_handle,
                         struct sockaddr_storage ip,
                         int flags,
                         TS_Attributes attributes,
                         TS_RatingSession *session,
                         TS_RatingCompleteFunc callback,
                         void *callback_data);

TS_API_PUBLIC int
TS_RateIPForSpam(TS_Handle ts_handle,
                 struct sockaddr_storage ip,
                 int flags,
                 TS_Attributes attributes);

#endif /* TS_FEATURE_FUNCTION_DECLARATIONS */

#ifdef __cplusplus
}
#endif/*  __cplusplus */


#endif /* TS_WEB_H */
