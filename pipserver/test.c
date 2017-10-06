#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "test.h"
#include "ts.h"

static TS_Handle TsHandle=NULL;

static char *rate_url(TS_Handle ts_handle, const char *url, int verbosity)
{
     char *ts_result=NULL;

     TS_Url parsed_url;
     TS_Attributes attributes;
     TS_Categories categories;
     int webrep=0;

     char cat_names[4000];
     int len = 0;
     char delimiter[] = ",";
     int delimiter_len = strlen(delimiter);
     int num_cats;
     unsigned int *cat_array=NULL;
     int i;

     if (TS_OK != TS_AttributesCreate(ts_handle, &attributes)) {
	  printf("TS_AttributesCreate failed. Abort.\n");
	  TS_HandleDestroy(&ts_handle);
	  return ts_result;
     }
     // printf("After TS_AttributesCreate()\n");

     if (TS_OK != TS_CategoriesCreate(ts_handle, &categories)) {
	  printf("TS_CategoriesCreate failed. Abort.\n");
	  TS_AttributesDestroy(ts_handle, &attributes);
	  TS_HandleDestroy(&ts_handle);
	  return ts_result;
     }
     // printf("After TS_CategoriesCreate()\n");

     if (TS_OK != TS_CategoriesCategoryRemoveAll(ts_handle, categories)) {
	  printf("TS_CategoriesCategoryRemoveAll failed. Abort.\n");
	  TS_AttributesDestroy(ts_handle, &attributes);
	  TS_CategoriesDestroy(ts_handle, &categories);
	  TS_HandleDestroy(&ts_handle);
	  return ts_result;
     }
     // printf("After TS_CategoriesCategoryRemoveAll()\n");

     if (TS_OK != TS_UrlCreate(ts_handle, &parsed_url)) {
	  printf("TS_UrlCreate failed. Abort.\n");
	  TS_AttributesDestroy(ts_handle, &attributes);
	  TS_CategoriesDestroy(ts_handle, &categories);
	  TS_HandleDestroy(&ts_handle);
	  return ts_result;
     }
     // printf("After TS_UrlCreate()\n");

     if (TS_OK != TS_UrlParse(ts_handle,
			      url,
			      NULL,
			      parsed_url)) {
	  printf("TS_UrlParse failed. Abort.\n");
	  goto done;
     }
     // printf("After TS_UrlParse()\n");

     if (TS_OK != TS_RateUrl(
	      ts_handle,
	      parsed_url,
	      attributes,
	      categories,
	      NULL,
	      0,
	      TS_CAT_SET_LOADED,
	      0,
	      NULL)) {
	  printf("TS_RateUrl failed. Abort.\n");
	  goto done;
     }
     // printf("After TS_RateUrl()\n");
/*
     if (TS_OK != TS_AttributesInfoGet(ts_handle,
				       attributes,
				       TS_ATTRIBUTES_INFO_REPUTATION,
				       &webrep,
				       sizeof(webrep))) {
	  printf("ERROR: cannot get web reputation!\n");
     } else {
	  printf("Webrep is %d\n", webrep);
     }
*/
     // Get categories number
     if (TS_OK != TS_CategoriesCount(ts_handle, categories, &num_cats)) {
	  printf("Get categories number error!\n");
     }
     // printf("After TS_CategoriesCount()\n");

     cat_array = (unsigned int*)malloc(num_cats+1);
     // Get categories codes
     for (i=0; i<num_cats+1; i++)
	  cat_array[i] = 0;

     // ignore errors
     if (TS_OK != TS_CategoriesToArray(ts_handle, categories, cat_array, &num_cats) ) {
	  printf("Failed to get category code array!\n");
     }
     // printf("After TS_CategoriesToArray()\n");


     len = sizeof(cat_names) - 1;
     if (TS_OK != TS_CategoriesToString(
	      ts_handle,
	      categories,
	      TS_LANGUAGE_ENGLISH,
	      TS_ENCODING_UTF8,
	      delimiter,
	      delimiter_len,
	      cat_names,
	      &len)) {
	  printf("TS_CategoriesToString failed. Abort.\n");
	  goto done;
     } else {
	  cat_names[len] = '\0';
	  if (strlen(cat_names) <= 1) {
	       if (verbosity >= 1)
    		    printf("x URL: '%s' is uncategorized!\n", url);
	  }
	  else {
	       ts_result = strdup(cat_names);
	       if (verbosity >= 2) {
		    //
		    printf("URL: '%s' is categorized as :'%s'\n", url, cat_names);
/*
                    printf("Category Codes: \n");
		    for (i=0; i<num_cats; i++) {
			 printf(" %u ", cat_array[i]);
		    }
*/
		    printf("\n");
	       }
	  }
     }
     // printf("After TS_CategoriesToString()\n");

done:
     if (cat_array != NULL) {
	  free(cat_array);
     }
     TS_AttributesDestroy(ts_handle, &attributes);
     TS_CategoriesDestroy(ts_handle, &categories);
     TS_UrlDestroy(ts_handle, &parsed_url);
     //TS_HandleDestroy(&ts_handle);

     return ts_result;
}


int Init() {
  TS_Database_Access db_access_mode;
  const char *returned_serial = NULL;
  const char *errors = NULL;
  const char *client_cert = NULL;
  const char *client_key = NULL;
  const char *trustedsource_server_cert = NULL;
  
  if (TS_Init() != TS_OK)
  {
    fprintf(stderr, "TS_Init Failed. Abort.\n");
    return 0;
  }
  
  if (TS_OK != TS_HandleCreate(
	   &TsHandle,
	   "SF6S-HH37-G34G-X75H",
	   NULL,
	   "Infoblox",
	   "1"))
  {
       fprintf(stderr, "TS_HandleCreate failed. Abort.\n");
       return 0;
  }

  if (TS_OK != TS_ActivateTrustedSource(
	   TsHandle,
	   TS_ACTIVATION_SERVER_DEFAULT,
	   NULL,
	   NULL,
	   &returned_serial,
	   &client_cert,
	   &client_key,
	   &trustedsource_server_cert,
	   &errors))
  {
       if (NULL == errors)
       {
	 fprintf(stderr, "Error during activation\n");
       }
       else
       {
	 fprintf(stderr, "Error from server: %s\n", errors);
       }
  }

  /*  db_access_mode = TS_DATABASE_ACCESS_MEMORY; */
  if (TS_OK != TS_DatabaseLoad( TsHandle,
				"data.db",
				TS_DATABASE_ACCESS_MEMORY,
				TS_CAT_SET_LATEST))
  {
       fprintf(stderr, "TS_DatabaseDownload failed. Abort!\n");
       TS_HandleDestroy(&TsHandle);
       return 0;
  }
  else
  {
       printf("Local DB is loaded successfully.\n");
  }

   return 0;
}


/*
 rate_url() returns a malloc'ed string that
 must be free'ed in calling code
*/
char *RateUrl(const char *url) {
    return rate_url(TsHandle, url, 2);
}

/*
int main(int argc, char *argv[]) {
     Init();
     RateUrl("www.thesun.co.uk");
}
*/
