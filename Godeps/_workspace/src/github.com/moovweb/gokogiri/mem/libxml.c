#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <time.h>

#include <libxml/xmlmemory.h>

//#define TRACE_MEM
//#define CUSTOM_GC

unsigned long alloc_count = 0;

#ifndef strdup
char *strdup (const char *str) {
	char *new = malloc(strlen(str));
	strcpy(new, str);
	return new;
}
#endif

#ifdef CUSTOM_GC
#pragma pack(push)
#pragma pack(1)
typedef struct go_xml_allocation {
	size_t size;
	struct timespec timestamp;
	void *p;
} go_xml_allocation;
#pragma pack(pop)
#endif

unsigned long libxmlGoAllocSize() {
	if (alloc_count > 0) {
		xmlCleanupParser();
	}
	return alloc_count;
}

void libxmlGoFree(void *p) {
	alloc_count--;
#ifdef CUSTOM_GC
	go_xml_allocation *gxa = (go_xml_allocation *)(p - sizeof(go_xml_allocation));
	fprintf(stderr, "Freeing %lu bytes @ %p created at: %lu\n", gxa->size, gxa->p, gxa->timestamp.tv_nsec);
	return free(gxa);
#else
#ifdef TRACE_MEM
	fprintf(stderr, "%08lu Free %p\n", alloc_count, p);
#endif
	return free(p);
#endif
}

void *libxmlGoMalloc(int size) {
	alloc_count++;
#ifdef CUSTOM_GC
	go_xml_allocation *gxa = (go_xml_allocation *)malloc(size + sizeof(go_xml_allocation));
	gxa->p = (void *)gxa + sizeof(go_xml_allocation);
	gxa->size = size;
	clock_gettime(CLOCK_REALTIME, &(gxa->timestamp));
	fprintf(stderr, "Allocated %lu bytes @ %p timestamp: %lu\n", gxa->size, gxa->p, gxa->timestamp.tv_nsec);
	return gxa->p;
#else
#ifdef TRACE_MEM
	fprintf(stderr, "%08lu Malloc %d\n", alloc_count, size);
#endif
	return malloc(size);
#endif
}

void *libxmlGoRealloc(void *p, int size) {
#ifdef TRACE_MEM
	fprintf(stderr, "Realloc %p, %d\n", p, size);
#endif
	return realloc(p, size);
}

void *libxmlGoStrDup(void *p) {
	alloc_count++;
#ifdef TRACE_MEM
	fprintf(stderr, "%08lu StrDup %p\n", alloc_count, p);
#endif
	return strdup(p);
}

void libxmlGoInit() {
#ifndef WINDOWS
	//fprintf(stderr, "Running xmlMemSetup()...\n");
	xmlMemSetup(
		(xmlFreeFunc)libxmlGoFree, 
		(xmlMallocFunc)libxmlGoMalloc, 
		(xmlReallocFunc)libxmlGoRealloc,
      	(xmlStrdupFunc)libxmlGoStrDup
	);
#endif

	//char *_LIBXML_VERSION = strdup(LIBXML_DOTTED_VERSION);
	//char *_LIBXML_PARSER_VERSION = strdup(xmlParserVersion);
	//fprintf(stderr, "LIBXML_VERSION: %s\n", _LIBXML_VERSION);
	//fprintf(stderr, "LIBXML_PARSER_VERSION: %s\n", _LIBXML_PARSER_VERSION);

#ifdef LIBXML_ICONV_ENABLED
	//fprintf(stderr, "LIBXML_ICONV_ENABLED: %s\n", "true");
#else
	//fprintf(stderr, "LIBXML_ICONV_ENABLED: %s\n", "false");
#endif

	//xmlInitParser();
}

