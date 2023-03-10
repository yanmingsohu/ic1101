/*
 * DMI Decode
 *
 *   Copyright (C) 2000-2002 Alan Cox <alan@redhat.com>
 *   Copyright (C) 2002-2008 Jean Delvare <khali@linux-fr.org>
 *
 *   This program is free software; you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation; either version 2 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program; if not, write to the Free Software
 *   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA  02111-1307 USA
 *
 *   For the avoidance of doubt the "preferred form" of this code is one which
 *   is in an open unpatent encumbered format. Where cryptographic key signing
 *   forms part of the process of creating an executable the information
 *   including keys needed to generate an equivalently functional executable
 *   are deemed to be part of the source code.
 *
 * Unless specified otherwise, all references are aimed at the "System
 * Management BIOS Reference Specification, Version 2.6" document,
 * available from http://www.dmtf.org/standards/smbios/.
 *
 * Note to contributors:
 * Please reference every value you add or modify, especially if the
 * information does not come from the above mentioned specification.
 *
 * Additional references:
 *  - Intel AP-485 revision 32
 *    "Intel Processor Identification and the CPUID Instruction"
 *    http://developer.intel.com/design/xeon/applnots/241618.htm
 *  - DMTF Common Information Model
 *    CIM Schema version 2.19.1
 *    http://www.dmtf.org/standards/cim/
 *  - IPMI 2.0 revision 1.0
 *    "Intelligent Platform Management Interface Specification"
 *    http://developer.intel.com/design/servers/ipmi/spec.htm
 *  - AMD publication #25481 revision 2.28
 *    "CPUID Specification"
 *    http://www.amd.com/us-en/assets/content_type/white_papers_and_tech_docs/25481.pdf
 *  - BIOS Integrity Services Application Programming Interface version 1.0
 *    http://www.intel.com/design/archives/wfm/downloads/bisspec.htm
 */

#include "fix_plat.h"
#include <stdio.h>
#include <string.h>
// #include <strings.h>
#include <stdlib.h>
// #include <unistd.h>

#include "version.h"
#include "config.h"
#include "types.h"
#include "util.h"
#include "dmidecode.h"
#include "dmiopt.h"
#include "dmioem.h"
 
#include "winsmbios.h"

#define out_of_spec "<OUT OF SPEC>"
static const char *bad_index = "<BAD INDEX>";
struct opt opt;

void fprintf2(void *a, ...) {
}

int printf2(const char *a, ...) {
	return 1;
}


/*
 * Type-independant Stuff
 */

const char *dmi_string(const struct dmi_header *dm, u8 s)
{
	char *bp = (char *)dm->data;
	size_t i, len;

	if (s == 0)
		return "Not Specified";

	bp += dm->length;
	while (s > 1 && *bp)
	{
		bp += strlen(bp);
		bp++;
		s--;
	}

	if (!*bp)
		return bad_index;

	if (!(opt.flags & FLAG_DUMP))
	{
		/* ASCII filtering */
		len = strlen(bp);
		for (i = 0; i < len; i++)
			if (bp[i] < 32 || bp[i] == 127)
				bp[i] = '.';
	}

	return bp;
}

static const char *dmi_smbios_structure_type(u8 code)
{
	static const char *type[] = {
		"BIOS", /* 0 */
		"System",
		"Base Board",
		"Chassis",
		"Processor",
		"Memory Controller",
		"Memory Module",
		"Cache",
		"Port Connector",
		"System Slots",
		"On Board Devices",
		"OEM Strings",
		"System Configuration Options",
		"BIOS Language",
		"Group Associations",
		"System Event Log",
		"Physical Memory Array",
		"Memory Device",
		"32-bit Memory Error",
		"Memory Array Mapped Address",
		"Memory Device Mapped Address",
		"Built-in Pointing Device",
		"Portable Battery",
		"System Reset",
		"Hardware Security",
		"System Power Controls",
		"Voltage Probe",
		"Cooling Device",
		"Temperature Probe",
		"Electrical Current Probe",
		"Out-of-band Remote Access",
		"Boot Integrity Services",
		"System Boot",
		"64-bit Memory Error",
		"Management Device",
		"Management Device Component",
		"Management Device Threshold Data",
		"Memory Channel",
		"IPMI Device",
		"Power Supply",
		"Additional Information",
		"Onboard Device", /* 41 */
	};

	if (code <= 41)
		return type[code];
	return out_of_spec;
}

static int dmi_bcd_range(u8 value, u8 low, u8 high)
{
	if (value > 0x99 || (value & 0x0F) > 0x09)
		return 0;
	if (value < low || value > high)
		return 0;
	return 1;
}

static void dmi_dump(const struct dmi_header *h, const char *prefix)
{
	int row, i;
	const char *s;

	printf2("%sHeader and Data:\n", prefix);
	for (row = 0; row < ((h->length - 1) >> 4) + 1; row++)
	{
		printf2("%s\t", prefix);
		for (i = 0; i < 16 && i < h->length - (row << 4); i++)
			printf2("%s%02X", i ? " " : "",
			       (h->data)[(row << 4) + i]);
		printf2("\n");
	}

	if ((h->data)[h->length] || (h->data)[h->length + 1])
	{
		printf2("%sStrings:\n", prefix);
		i = 1;
		while ((s = dmi_string(h, i++)) != bad_index)
		{
			if (opt.flags & FLAG_DUMP)
			{
				int j, l = strlen(s) + 1;
				for (row = 0; row < ((l - 1) >> 4) + 1; row++)
				{
					printf2("%s\t", prefix);
					for (j = 0; j < 16 && j < l - (row << 4); j++)
						printf2("%s%02X", j ? " " : "",
						       s[(row << 4) + j]);
					printf2("\n");
				}
				/* String isn't filtered yet so do it now */
				printf2("%s\t\"", prefix);
				while (*s)
				{
					if (*s < 32 || *s == 127)
						fputc('.', stdout);
					else
						fputc(*s, stdout);
					s++;
				}
				printf2("\"\n");
			}
			else
				printf2("%s\t%s\n", prefix, s);
		}
	}
}

/*
 * 3.3.1 BIOS Information (Type 0)
 */

static void dmi_bios_runtime_size(u32 code)
{
	if (code & 0x000003FF)
		printf2(" %u bytes", code);
	else
		printf2(" %u kB", code >> 10);
}

static void dmi_bios_characteristics(u64 code, const char *prefix)
{
	/* 3.3.1.1 */
	static const char *characteristics[] = {
		"BIOS characteristics not supported", /* 3 */
		"ISA is supported",
		"MCA is supported",
		"EISA is supported",
		"PCI is supported",
		"PC Card (PCMCIA) is supported",
		"PNP is supported",
		"APM is supported",
		"BIOS is upgradeable",
		"BIOS shadowing is allowed",
		"VLB is supported",
		"ESCD support is available",
		"Boot from CD is supported",
		"Selectable boot is supported",
		"BIOS ROM is socketed",
		"Boot from PC Card (PCMCIA) is supported",
		"EDD is supported",
		"Japanese floppy for NEC 9800 1.2 MB is supported (int 13h)",
		"Japanese floppy for Toshiba 1.2 MB is supported (int 13h)",
		"5.25\"/360 kB floppy services are supported (int 13h)",
		"5.25\"/1.2 MB floppy services are supported (int 13h)",
		"3.5\"/720 kB floppy services are supported (int 13h)",
		"3.5\"/2.88 MB floppy services are supported (int 13h)",
		"Print screen service is supported (int 5h)",
		"8042 keyboard services are supported (int 9h)",
		"Serial services are supported (int 14h)",
		"Printer services are supported (int 17h)",
		"CGA/mono video services are supported (int 10h)",
		"NEC PC-98" /* 31 */
	};
	int i;

	/*
	 * This isn't very clear what this bit is supposed to mean
	 */
	if (code.l & (1 << 3))
	{
		printf2("%s%s\n",
			prefix, characteristics[0]);
		return;
	}

	for (i = 4; i <= 31; i++)
		if (code.l & (1 << i))
			printf2("%s%s\n",
				prefix, characteristics[i - 3]);
}

static void dmi_bios_characteristics_x1(u8 code, const char *prefix)
{
	/* 3.3.1.2.1 */
	static const char *characteristics[] = {
		"ACPI is supported", /* 0 */
		"USB legacy is supported",
		"AGP is supported",
		"I2O boot is supported",
		"LS-120 boot is supported",
		"ATAPI Zip drive boot is supported",
		"IEEE 1394 boot is supported",
		"Smart battery is supported" /* 7 */
	};
	int i;

	for (i = 0; i <= 7; i++)
		if (code & (1 << i))
			printf2("%s%s\n",
				prefix, characteristics[i]);
}

static void dmi_bios_characteristics_x2(u8 code, const char *prefix)
{
	/* 3.3.1.2.2 */
	static const char *characteristics[] = {
		"BIOS boot specification is supported", /* 0 */
		"Function key-initiated network boot is supported",
		"Targeted content distribution is supported" /* 2 */
	};
	int i;

	for (i = 0; i <= 2; i++)
		if (code & (1 << i))
			printf2("%s%s\n",
				prefix, characteristics[i]);
}

/*
 * 3.3.2 System Information (Type 1)
 */

static void dmi_system_uuid(const u8 *p, u16 ver)
{
	int only0xFF = 1, only0x00 = 1;
	int i;

	for (i = 0; i < 16 && (only0x00 || only0xFF); i++)
	{
		if (p[i] != 0x00) only0x00 = 0;
		if (p[i] != 0xFF) only0xFF = 0;
	}

	if (only0xFF)
	{
		printf2("Not Present");
		return;
	}
	if (only0x00)
	{
		printf2("Not Settable");
		return;
	}

	/*
	 * As off version 2.6 of the SMBIOS specification, the first 3
	 * fields of the UUID are supposed to be encoded on little-endian.
	 * The specification says that this is the defacto standard,
	 * however I've seen systems following RFC 4122 instead and use
	 * network byte order, so I am reluctant to apply the byte-swapping
	 * for older versions.
	 */
	if (ver >= 0x0206)
		printf2("%02X%02X%02X%02X-%02X%02X-%02X%02X-%02X%02X-%02X%02X%02X%02X%02X%02X",
			p[3], p[2], p[1], p[0], p[5], p[4], p[7], p[6],
			p[8], p[9], p[10], p[11], p[12], p[13], p[14], p[15]);
	else
		printf2("%02X%02X%02X%02X-%02X%02X-%02X%02X-%02X%02X-%02X%02X%02X%02X%02X%02X",
			p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7],
			p[8], p[9], p[10], p[11], p[12], p[13], p[14], p[15]);
}

static const char *dmi_system_wake_up_type(u8 code)
{
	/* 3.3.2.1 */
	static const char *type[] = {
		"Reserved", /* 0x00 */
		"Other",
		"Unknown",
		"APM Timer",
		"Modem Ring",
		"LAN Remote",
		"Power Switch",
		"PCI PME#",
		"AC Power Restored" /* 0x08 */
	};

	if (code <= 0x08)
		return type[code];
	return out_of_spec;
}

/*
 * 3.3.3 Base Board Information (Type 2)
 */

static void dmi_base_board_features(u8 code, const char *prefix)
{
	/* 3.3.3.1 */
	static const char *features[] = {
		"Board is a hosting board", /* 0 */
		"Board requires at least one daughter board",
		"Board is removable",
		"Board is replaceable",
		"Board is hot swappable" /* 4 */
	};

	if ((code & 0x1F) == 0)
		printf2(" None\n");
	else
	{
		int i;

		printf2("\n");
		for (i = 0; i <= 4; i++)
			if (code & (1 << i))
				printf2("%s%s\n",
					prefix, features[i]);
	}
}

static const char *dmi_base_board_type(u8 code)
{
	/* 3.3.3.2 */
	static const char *type[] = {
		"Unknown", /* 0x01 */
		"Other",
		"Server Blade",
		"Connectivity Switch",
		"System Management Module",
		"Processor Module",
		"I/O Module",
		"Memory Module",
		"Daughter Board",
		"Motherboard",
		"Processor+Memory Module",
		"Processor+I/O Module",
		"Interconnect Board" /* 0x0D */
	};

	if (code >= 0x01 && code <= 0x0D)
		return type[code - 0x01];
	return out_of_spec;
}

static void dmi_base_board_handles(u8 count, const u8 *p, const char *prefix)
{
	int i;

	printf2("%sContained Object Handles: %u\n",
		prefix, count);
	for (i = 0; i < count; i++)
		printf2("%s\t0x%04X\n",
			prefix, WORD(p + sizeof(u16) * i));
}

/*
 * 3.3.4 Chassis Information (Type 3)
 */

static const char *dmi_chassis_type(u8 code)
{
	/* 3.3.4.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Desktop",
		"Low Profile Desktop",
		"Pizza Box",
		"Mini Tower",
		"Tower",
		"Portable",
		"Laptop",
		"Notebook",
		"Hand Held",
		"Docking Station",
		"All In One",
		"Sub Notebook",
		"Space-saving",
		"Lunch Box",
		"Main Server Chassis", /* CIM_Chassis.ChassisPackageType says "Main System Chassis" */
		"Expansion Chassis",
		"Sub Chassis",
		"Bus Expansion Chassis",
		"Peripheral Chassis",
		"RAID Chassis",
		"Rack Mount Chassis",
		"Sealed-case PC",
		"Multi-system",
		"CompactPCI",
		"AdvancedTCA",
		"Blade",
		"Blade Enclosing" /* 0x1D */
	};

	if (code >= 0x01 && code <= 0x1D)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_chassis_lock(u8 code)
{
	static const char *lock[] = {
		"Not Present", /* 0x00 */
		"Present" /* 0x01 */
	};

	return lock[code];
}

static const char *dmi_chassis_state(u8 code)
{
	/* 3.3.4.2 */
	static const char *state[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Safe",
		"Warning",
		"Critical",
		"Non-recoverable" /* 0x06 */
	};

	if (code >= 0x01 && code <= 0x06)
		return state[code - 0x01];
	return out_of_spec;
}

static const char *dmi_chassis_security_status(u8 code)
{
	/* 3.3.4.3 */
	static const char *status[] = {
		"Other", /* 0x01 */
		"Unknown",
		"None",
		"External Interface Locked Out",
		"External Interface Enabled" /* 0x05 */
	};

	if (code >= 0x01 && code <= 0x05)
		return status[code - 0x01];
	return out_of_spec;
}

static void dmi_chassis_height(u8 code)
{
	if (code == 0x00)
		printf2(" Unspecified");
	else
		printf2(" %u U", code);
}

static void dmi_chassis_power_cords(u8 code)
{
	if (code == 0x00)
		printf2(" Unspecified");
	else
		printf2(" %u", code);
}

static void dmi_chassis_elements(u8 count, u8 len, const u8 *p, const char *prefix)
{
	int i;

	printf2("%sContained Elements: %u\n",
		prefix, count);
	for (i = 0; i < count; i++)
	{
		if (len >= 0x03)
		{
			printf2("%s\t%s (",
				prefix, p[i * len] & 0x80 ?
				dmi_smbios_structure_type(p[i * len] & 0x7F) :
				dmi_base_board_type(p[i * len] & 0x7F));
			if (p[1 + i * len] == p[2 + i * len])
				printf2("%u", p[1 + i * len]);
			else
				printf2("%u-%u", p[1 + i * len], p[2 + i * len]);
			printf2(")\n");
		}
	}
}

/*
 * 3.3.5 Processor Information (Type 4)
 */

static const char *dmi_processor_type(u8 code)
{
	/* 3.3.5.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Central Processor",
		"Math Processor",
		"DSP Processor",
		"Video Processor" /* 0x06 */
	};

	if (code >= 0x01 && code <= 0x06)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_processor_family(const struct dmi_header *h)
{
	const u8 *data = h->data;
	unsigned int i, low, high;
	u16 code;

	/* 3.3.5.2 */
	static struct {
		int value;
		const char *name;
	} family2[] = {
		{ 0x01, "Other" },
		{ 0x02, "Unknown" },
		{ 0x03, "8086" },
		{ 0x04, "80286" },
		{ 0x05, "80386" },
		{ 0x06, "80486" },
		{ 0x07, "8087" },
		{ 0x08, "80287" },
		{ 0x09, "80387" },
		{ 0x0A, "80487" },
		{ 0x0B, "Pentium" },
		{ 0x0C, "Pentium Pro" },
		{ 0x0D, "Pentium II" },
		{ 0x0E, "Pentium MMX" },
		{ 0x0F, "Celeron" },
		{ 0x10, "Pentium II Xeon" },
		{ 0x11, "Pentium III" },
		{ 0x12, "M1" },
		{ 0x13, "M2" },
		{ 0x14, "Celeron M" }, /* From CIM_Processor.Family */
		{ 0x15, "Pentium 4 HT" }, /* From CIM_Processor.Family */

		{ 0x18, "Duron" },
		{ 0x19, "K5" },
		{ 0x1A, "K6" },
		{ 0x1B, "K6-2" },
		{ 0x1C, "K6-3" },
		{ 0x1D, "Athlon" },
		{ 0x1E, "AMD29000" },
		{ 0x1F, "K6-2+" },
		{ 0x20, "Power PC" },
		{ 0x21, "Power PC 601" },
		{ 0x22, "Power PC 603" },
		{ 0x23, "Power PC 603+" },
		{ 0x24, "Power PC 604" },
		{ 0x25, "Power PC 620" },
		{ 0x26, "Power PC x704" },
		{ 0x27, "Power PC 750" },
		{ 0x28, "Core Duo" }, /* From CIM_Processor.Family */
		{ 0x29, "Core Duo Mobile" }, /* From CIM_Processor.Family */
		{ 0x2A, "Core Solo Mobile" }, /* From CIM_Processor.Family */
		{ 0x2B, "Atom" }, /* From CIM_Processor.Family */

		{ 0x30, "Alpha" },
		{ 0x31, "Alpha 21064" },
		{ 0x32, "Alpha 21066" },
		{ 0x33, "Alpha 21164" },
		{ 0x34, "Alpha 21164PC" },
		{ 0x35, "Alpha 21164a" },
		{ 0x36, "Alpha 21264" },
		{ 0x37, "Alpha 21364" },

		{ 0x40, "MIPS" },
		{ 0x41, "MIPS R4000" },
		{ 0x42, "MIPS R4200" },
		{ 0x43, "MIPS R4400" },
		{ 0x44, "MIPS R4600" },
		{ 0x45, "MIPS R10000" },

		{ 0x50, "SPARC" },
		{ 0x51, "SuperSPARC" },
		{ 0x52, "MicroSPARC II" },
		{ 0x53, "MicroSPARC IIep" },
		{ 0x54, "UltraSPARC" },
		{ 0x55, "UltraSPARC II" },
		{ 0x56, "UltraSPARC IIi" },
		{ 0x57, "UltraSPARC III" },
		{ 0x58, "UltraSPARC IIIi" },

		{ 0x60, "68040" },
		{ 0x61, "68xxx" },
		{ 0x62, "68000" },
		{ 0x63, "68010" },
		{ 0x64, "68020" },
		{ 0x65, "68030" },

		{ 0x70, "Hobbit" },

		{ 0x78, "Crusoe TM5000" },
		{ 0x79, "Crusoe TM3000" },
		{ 0x7A, "Efficeon TM8000" },

		{ 0x80, "Weitek" },

		{ 0x82, "Itanium" },
		{ 0x83, "Athlon 64" },
		{ 0x84, "Opteron" },
		{ 0x85, "Sempron" },
		{ 0x86, "Turion 64" },
		{ 0x87, "Dual-Core Opteron" },
		{ 0x88, "Athlon 64 X2" },
		{ 0x89, "Turion 64 X2" },
		{ 0x8A, "Quad-Core Opteron" }, /* From CIM_Processor.Family */
		{ 0x8B, "Third-Generation Opteron" }, /* From CIM_Processor.Family */
		{ 0x8C, "Phenom FX" }, /* From CIM_Processor.Family */
		{ 0x8D, "Phenom X4" }, /* From CIM_Processor.Family */
		{ 0x8E, "Phenom X2" }, /* From CIM_Processor.Family */
		{ 0x8F, "Athlon X2" }, /* From CIM_Processor.Family */
		{ 0x90, "PA-RISC" },
		{ 0x91, "PA-RISC 8500" },
		{ 0x92, "PA-RISC 8000" },
		{ 0x93, "PA-RISC 7300LC" },
		{ 0x94, "PA-RISC 7200" },
		{ 0x95, "PA-RISC 7100LC" },
		{ 0x96, "PA-RISC 7100" },

		{ 0xA0, "V30" },
		{ 0xA1, "Quad-Core Xeon 3200" }, /* From CIM_Processor.Family */
		{ 0xA2, "Dual-Core Xeon 3000" }, /* From CIM_Processor.Family */
		{ 0xA3, "Quad-Core Xeon 5300" }, /* From CIM_Processor.Family */
		{ 0xA4, "Dual-Core Xeon 5100" }, /* From CIM_Processor.Family */
		{ 0xA5, "Dual-Core Xeon 5000" }, /* From CIM_Processor.Family */
		{ 0xA6, "Dual-Core Xeon LV" }, /* From CIM_Processor.Family */
		{ 0xA7, "Dual-Core Xeon ULV" }, /* From CIM_Processor.Family */
		{ 0xA8, "Dual-Core Xeon 7100" }, /* From CIM_Processor.Family */
		{ 0xA9, "Quad-Core Xeon 5400" }, /* From CIM_Processor.Family */
		{ 0xAA, "Quad-Core Xeon" }, /* From CIM_Processor.Family */

		{ 0xB0, "Pentium III Xeon" },
		{ 0xB1, "Pentium III Speedstep" },
		{ 0xB2, "Pentium 4" },
		{ 0xB3, "Xeon" },
		{ 0xB4, "AS400" },
		{ 0xB5, "Xeon MP" },
		{ 0xB6, "Athlon XP" },
		{ 0xB7, "Athlon MP" },
		{ 0xB8, "Itanium 2" },
		{ 0xB9, "Pentium M" },
		{ 0xBA, "Celeron D" },
		{ 0xBB, "Pentium D" },
		{ 0xBC, "Pentium EE" },
		{ 0xBD, "Core Solo" },
		/* 0xBE handled as a special case */
		{ 0xBF, "Core 2 Duo" },
		{ 0xC0, "Core 2 Solo" }, /* From CIM_Processor.Family */
		{ 0xC1, "Core 2 Extreme" }, /* From CIM_Processor.Family */
		{ 0xC2, "Core 2 Quad" }, /* From CIM_Processor.Family */
		{ 0xC3, "Core 2 Extreme Mobile" }, /* From CIM_Processor.Family */
		{ 0xC4, "Core 2 Duo Mobile" }, /* From CIM_Processor.Family */
		{ 0xC5, "Core 2 Solo Mobile" }, /* From CIM_Processor.Family */

		{ 0xC8, "IBM390" },
		{ 0xC9, "G4" },
		{ 0xCA, "G5" },
		{ 0xCB, "ESA/390 G6" },
		{ 0xCC, "z/Architectur" },

		{ 0xD2, "C7-M" },
		{ 0xD3, "C7-D" },
		{ 0xD4, "C7" },
		{ 0xD5, "Eden" },

		{ 0xE6, "Embedded Opteron Quad-Core" }, /* From CIM_Processor.Family */
		{ 0xE7, "Phenom Triple-Core" }, /* From CIM_Processor.Family */
		{ 0xE8, "Turion Ultra Dual-Core Mobile" }, /* From CIM_Processor.Family */
		{ 0xE9, "Turion Dual-Core Mobile" }, /* From CIM_Processor.Family */
		{ 0xEA, "Athlon Dual-Core" }, /* From CIM_Processor.Family */
		{ 0xEB, "Sempron SI" }, /* From CIM_Processor.Family */

		{ 0xFA, "i860" },
		{ 0xFB, "i960" },

		{ 0x104, "SH-3" },
		{ 0x105, "SH-4" },
		{ 0x118, "ARM" },
		{ 0x119, "StrongARM" },
		{ 0x12C, "6x86" },
		{ 0x12D, "MediaGX" },
		{ 0x12E, "MII" },
		{ 0x140, "WinChip" },
		{ 0x15E, "DSP" },
		{ 0x1F4, "Video Processor" },
	};

	code = (data[0x06] == 0xFE && h->length >= 0x2A) ?
		WORD(data + 0x28) : data[0x06];

	/* Special case for ambiguous value 0xBE */
	if (code == 0xBE)
	{
		const char *manufacturer = dmi_string(h, data[0x07]);

		/* Best bet based on manufacturer string */
		if (strstr(manufacturer, "Intel") != NULL
		 || strncmp(manufacturer, "Intel", 5) == 0)
			return "Core 2";
		if (strstr(manufacturer, "AMD") != NULL
		 || strncmp(manufacturer, "AMD", 3) == 0)
			return "K7";
		return "Core 2 or K7";
	}

	/* Perform a binary search */
	low = 0;
	high = ARRAY_SIZE(family2) - 1;

	while (1)
	{
		i = (low + high) / 2;
		if (family2[i].value == code)
			return family2[i].name;
		if (low == high) /* Not found */
			return out_of_spec;

		if (code < family2[i].value)
			high = i;
		else
			low = i + 1;
	}
}

static void dmi_processor_id(u8 type, const u8 *p, const char *version, const char *prefix)
{
	/* Intel AP-485 revision 32, table 3-4 */
	static const char *flags[32] = {
		"FPU (Floating-point unit on-chip)", /* 0 */
		"VME (Virtual mode extension)",
		"DE (Debugging extension)",
		"PSE (Page size extension)",
		"TSC (Time stamp counter)",
		"MSR (Model specific registers)",
		"PAE (Physical address extension)",
		"MCE (Machine check exception)",
		"CX8 (CMPXCHG8 instruction supported)",
		"APIC (On-chip APIC hardware supported)",
		NULL, /* 10 */
		"SEP (Fast system call)",
		"MTRR (Memory type range registers)",
		"PGE (Page global enable)",
		"MCA (Machine check architecture)",
		"CMOV (Conditional move instruction supported)",
		"PAT (Page attribute table)",
		"PSE-36 (36-bit page size extension)",
		"PSN (Processor serial number present and enabled)",
		"CLFSH (CLFLUSH instruction supported)",
		NULL, /* 20 */
		"DS (Debug store)",
		"ACPI (ACPI supported)",
		"MMX (MMX technology supported)",
		"FXSR (Fast floating-point save and restore)",
		"SSE (Streaming SIMD extensions)",
		"SSE2 (Streaming SIMD extensions 2)",
		"SS (Self-snoop)",
		"HTT (Hyper-threading technology)",
		"TM (Thermal monitor supported)",
		"IA64 (IA64 capabilities)",
		"PBE (Pending break enabled)" /* 31 */
	};
	/*
	 * Extra flags are now returned in the ECX register when one calls
	 * the CPUID instruction. Their meaning is explained in table 3-5, but
	 * DMI doesn't support this yet.
	 */
	u32 eax, edx;
	int sig = 0;

	/*
	 * This might help learn about new processors supporting the
	 * CPUID instruction or another form of identification.
	 */
	printf2("%sID: %02X %02X %02X %02X %02X %02X %02X %02X\n",
		prefix, p[0], p[1], p[2], p[3], p[4], p[5], p[6], p[7]);

	if (type == 0x05) /* 80386 */
	{
		u16 dx = WORD(p);
		/*
		 * 80386 have a different signature.
		 */
		printf2("%sSignature: Type %u, Family %u, Major Stepping %u, Minor Stepping %u\n",
			prefix, dx >> 12, (dx >> 8) & 0xF,
			(dx >> 4) & 0xF, dx & 0xF);
		return;
	}
	if (type == 0x06) /* 80486 */
	{
		u16 dx = WORD(p);
		/*
		 * Not all 80486 CPU support the CPUID instruction, we have to find
		 * wether the one we have here does or not. Note that this trick
		 * works only because we know that 80486 must be little-endian.
		 */
		if ((dx & 0x0F00) == 0x0400
		 && ((dx & 0x00F0) == 0x0040 || (dx & 0x00F0) >= 0x0070)
		 && ((dx & 0x000F) >= 0x0003))
			sig = 1;
		else
		{
			printf2("%sSignature: Type %u, Family %u, Model %u, Stepping %u\n",
				prefix, (dx >> 12) & 0x3, (dx >> 8) & 0xF,
				(dx >> 4) & 0xF, dx & 0xF);
			return;
		}
	}
	else if ((type >= 0x0B && type <= 0x15) /* Intel, Cyrix */
	      || (type >= 0x28 && type <= 0x2B) /* Intel */
	      || (type >= 0xA1 && type <= 0xAA) /* Intel */
	      || (type >= 0xB0 && type <= 0xB3) /* Intel */
	      || type == 0xB5 /* Intel */
	      || (type >= 0xB9 && type <= 0xC5) /* Intel */
	      || (type >= 0xD2 && type <= 0xD5)) /* VIA */
		sig = 1;
	else if ((type >= 0x18 && type <= 0x1D) /* AMD */
	      || type == 0x1F /* AMD */
	      || (type >= 0x83 && type <= 0x8F) /* AMD */
	      || (type >= 0xB6 && type <= 0xB7) /* AMD */
	      || (type >= 0xE6 && type <= 0xEB)) /* AMD */
		sig = 2;
	else if (type == 0x01 || type == 0x02)
	{
		/*
		 * Some X86-class CPU have family "Other" or "Unknown". In this case,
		 * we use the version string to determine if they are known to
		 * support the CPUID instruction.
		 */
		if (strncmp(version, "Pentium III MMX", 15) == 0
		 || strncmp(version, "Intel(R) Core(TM)2", 18) == 0
		 || strncmp(version, "Intel(R) Pentium(R)", 19) == 0
		 || strcmp(version, "Genuine Intel(R) CPU U1400") == 0)
			sig = 1;
		else if (strncmp(version, "AMD Athlon(TM)", 14) == 0
		      || strncmp(version, "AMD Opteron(tm)", 15) == 0
		      || strncmp(version, "Dual-Core AMD Opteron(tm)", 25) == 0)
			sig = 2;
		else
			return;
	}
	else /* not X86-class */
		return;

	eax = DWORD(p);
	edx = DWORD(p + 4);
	switch (sig)
	{
		case 1: /* Intel */
			printf2("%sSignature: Type %u, Family %u, Model %u, Stepping %u\n",
				prefix, (eax >> 12) & 0x3,
				((eax >> 20) & 0xFF) + ((eax >> 8) & 0x0F),
				((eax >> 12) & 0xF0) + ((eax >> 4) & 0x0F),
				eax & 0xF);
			break;
		case 2: /* AMD, publication #25481 revision 2.28 */
			printf2("%sSignature: Family %u, Model %u, Stepping %u\n",
				prefix,
				((eax >> 8) & 0xF) + (((eax >> 8) & 0xF) == 0xF ? (eax >> 20) & 0xFF : 0),
				((eax >> 4) & 0xF) | (((eax >> 8) & 0xF) == 0xF ? (eax >> 12) & 0xF0 : 0),
				eax&0xF);
			break;
	}

	edx = DWORD(p + 4);
	printf2("%sFlags:", prefix);
	if ((edx & 0xFFEFFBFF) == 0)
		printf2(" None\n");
	else
	{
		int i;

		printf2("\n");
		for (i = 0; i <= 31; i++)
			if (flags[i] != NULL && edx & (1 << i))
				printf2("%s\t%s\n", prefix, flags[i]);
	}
}

static void dmi_processor_voltage(u8 code)
{
	/* 3.3.5.4 */
	static const char *voltage[] = {
		"5.0 V", /* 0 */
		"3.3 V",
		"2.9 V" /* 2 */
	};
	int i;

	if (code & 0x80)
		printf2(" %.1f V", (float)(code & 0x7f) / 10);
	else
	{
		for (i = 0; i <= 2; i++)
			if (code & (1 << i))
				printf2(" %s", voltage[i]);
		if (code == 0x00)
			printf2(" Unknown");
	}
}

static void dmi_processor_frequency(const u8 *p)
{
	u16 code = WORD(p);

	if (code)
		printf2("%u MHz", code);
	else
		printf2("Unknown");
}

/* code is assumed to be a 3-bit value */
static const char *dmi_processor_status(u8 code)
{
	static const char *status[] = {
		"Unknown", /* 0x00 */
		"Enabled",
		"Disabled By User",
		"Disabled By BIOS",
		"Idle", /* 0x04 */
		out_of_spec,
		out_of_spec,
		"Other" /* 0x07 */
	};

	return status[code];
}

static const char *dmi_processor_upgrade(u8 code)
{
	/* 3.3.5.5 */
	static const char *upgrade[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Daughter Board",
		"ZIF Socket",
		"Replaceable Piggy Back",
		"None",
		"LIF Socket",
		"Slot 1",
		"Slot 2",
		"370-pin Socket",
		"Slot A",
		"Slot M",
		"Socket 423",
		"Socket A (Socket 462)",
		"Socket 478",
		"Socket 754",
		"Socket 940",
		"Socket 939",
		"Socket mPGA604",
		"Socket LGA771",
		"Socket LGA775",
		"Socket S1",
		"Socket AM2",
		"Socket F (1207)" /* 0x18 */
	};

	if (code >= 0x01 && code <= 0x18)
		return upgrade[code - 0x01];
	return out_of_spec;
}

static void dmi_processor_cache(u16 code, const char *level, u16 ver)
{
	if (code == 0xFFFF)
	{
		if (ver >= 0x0203)
			printf2(" Not Provided");
		else
			printf2(" No %s Cache", level);
	}
	else
		printf2(" 0x%04X", code);
}

static void dmi_processor_characteristics(u16 code, const char *prefix)
{
	/* 3.3.5.9 */
	static const char *characteristics[] = {
		"64-bit capable" /* 2 */
	};

	if ((code & 0x0004) == 0)
		printf2(" None\n");
	else
	{
		int i;

		printf2("\n");
		for (i = 2; i <= 2; i++)
			if (code & (1 << i))
				printf2("%s%s\n", prefix, characteristics[i - 2]);
	}
}

/*
 * 3.3.6 Memory Controller Information (Type 5)
 */

static const char *dmi_memory_controller_ed_method(u8 code)
{
	/* 3.3.6.1 */
	static const char *method[] = {
		"Other", /* 0x01 */
		"Unknown",
		"None",
		"8-bit Parity",
		"32-bit ECC",
		"64-bit ECC",
		"128-bit ECC",
		"CRC" /* 0x08 */
	};

	if (code >= 0x01 && code <= 0x08)
		return method[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_controller_ec_capabilities(u8 code, const char *prefix)
{
	/* 3.3.6.2 */
	static const char *capabilities[] = {
		"Other", /* 0 */
		"Unknown",
		"None",
		"Single-bit Error Correcting",
		"Double-bit Error Correcting",
		"Error Scrubbing" /* 5 */
	};

	if ((code & 0x3F) == 0)
		printf2(" None\n");
	else
	{
		int i;

		printf2("\n");
		for (i = 0; i <= 5; i++)
			if (code & (1 << i))
				printf2("%s%s\n", prefix, capabilities[i]);
	}
}

static const char *dmi_memory_controller_interleave(u8 code)
{
	/* 3.3.6.3 */
	static const char *interleave[] = {
		"Other", /* 0x01 */
		"Unknown",
		"One-way Interleave",
		"Two-way Interleave",
		"Four-way Interleave",
		"Eight-way Interleave",
		"Sixteen-way Interleave" /* 0x07 */
	};

	if (code >= 0x01 && code <= 0x07)
		return interleave[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_controller_speeds(u16 code, const char *prefix)
{
	/* 3.3.6.4 */
	const char *speeds[] = {
		"Other", /* 0 */
		"Unknown",
		"70 ns",
		"60 ns",
		"50 ns" /* 4 */
	};

	if ((code & 0x001F) == 0)
		printf2(" None\n");
	else
	{
		int i;

		printf2("\n");
		for (i = 0; i <= 4; i++)
			if (code & (1 << i))
				printf2("%s%s\n", prefix, speeds[i]);
	}
}

static void dmi_memory_controller_slots(u8 count, const u8 *p, const char *prefix)
{
	int i;

	printf2("%sAssociated Memory Slots: %u\n",
		prefix, count);
	for (i = 0; i < count; i++)
		printf2("%s\t0x%04X\n",
			prefix, WORD(p + sizeof(u16) * i));
}

/*
 * 3.3.7 Memory Module Information (Type 6)
 */

static void dmi_memory_module_types(u16 code, const char *sep)
{
	/* 3.3.7.1 */
	static const char *types[] = {
		"Other", /* 0 */
		"Unknown",
		"Standard",
		"FPM",
		"EDO",
		"Parity",
		"ECC",
		"SIMM",
		"DIMM",
		"Burst EDO",
		"SDRAM" /* 10 */
	};

	if ((code & 0x07FF) == 0)
		printf2(" None");
	else
	{
		int i;

		for (i = 0; i <= 10; i++)
			if (code & (1 << i))
				printf2("%s%s", sep, types[i]);
	}
}

static void dmi_memory_module_connections(u8 code)
{
	if (code == 0xFF)
		printf2(" None");
	else
	{
		if ((code & 0xF0) != 0xF0)
			printf2(" %u", code >> 4);
		if ((code & 0x0F) != 0x0F)
			printf2(" %u", code & 0x0F);
	}
}

static void dmi_memory_module_speed(u8 code)
{
	if (code == 0)
		printf2(" Unknown");
	else
		printf2(" %u ns", code);
}

static void dmi_memory_module_size(u8 code)
{
	/* 3.3.7.2 */
	switch (code & 0x7F)
	{
		case 0x7D:
			printf2(" Not Determinable");
			break;
		case 0x7E:
			printf2(" Disabled");
			break;
		case 0x7F:
			printf2(" Not Installed");
			return;
		default:
			printf2(" %u MB", 1 << (code & 0x7F));
	}

	if (code & 0x80)
		printf2(" (Double-bank Connection)");
	else
		printf2(" (Single-bank Connection)");
}

static void dmi_memory_module_error(u8 code, const char *prefix)
{
	if (code & (1 << 2))
		printf2(" See Event Log\n");
	else
	{	if ((code & 0x03) == 0)
			printf2(" OK\n");
		if (code & (1 << 0))
			printf2("%sUncorrectable Errors\n", prefix);
		if (code & (1 << 1))
			printf2("%sCorrectable Errors\n", prefix);
	}
}

/*
 * 3.3.8 Cache Information (Type 7)
 */

static const char *dmi_cache_mode(u8 code)
{
	static const char *mode[] = {
		"Write Through", /* 0x00 */
		"Write Back",
		"Varies With Memory Address",
		"Unknown" /* 0x03 */
	};

	return mode[code];
}

/* code is assumed to be a 2-bit value */
static const char *dmi_cache_location(u8 code)
{
	static const char *location[4] = {
		"Internal", /* 0x00 */
		"External",
		out_of_spec, /* 0x02 */
		"Unknown" /* 0x03 */
	};

	return location[code];
}

static void dmi_cache_size(u16 code)
{
	if (code & 0x8000)
		printf2(" %u kB", (code & 0x7FFF) << 6);
	else
		printf2(" %u kB", code);
}

static void dmi_cache_types(u16 code, const char *sep)
{
	/* 3.3.8.2 */
	static const char *types[] = {
		"Other", /* 0 */
		"Unknown",
		"Non-burst",
		"Burst",
		"Pipeline Burst",
		"Synchronous",
		"Asynchronous" /* 6 */
	};

	if ((code & 0x007F) == 0)
		printf2(" None");
	else
	{
		int i;

		for (i = 0; i <= 6; i++)
			if (code & (1 << i))
				printf2("%s%s", sep, types[i]);
	}
}

static const char *dmi_cache_ec_type(u8 code)
{
	/* 3.3.8.3 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"None",
		"Parity",
		"Single-bit ECC",
		"Multi-bit ECC" /* 0x06 */
	};

	if (code >= 0x01 && code <= 0x06)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_cache_type(u8 code)
{
	/* 3.3.8.4 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Instruction",
		"Data",
		"Unified" /* 0x05 */
	};

	if (code >= 0x01 && code <= 0x05)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_cache_associativity(u8 code)
{
	/* 3.3.8.5 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Direct Mapped",
		"2-way Set-associative",
		"4-way Set-associative",
		"Fully Associative",
		"8-way Set-associative",
		"16-way Set-associative" /* 0x08 */
	};

	if (code >= 0x01 && code <= 0x08)
		return type[code - 0x01];
	return out_of_spec;
}

/*
 * 3.3.9 Port Connector Information (Type 8)
 */

static const char *dmi_port_connector_type(u8 code)
{
	/* 3.3.9.2 */
	static const char *type[] = {
		"None", /* 0x00 */
		"Centronics",
		"Mini Centronics",
		"Proprietary",
		"DB-25 male",
		"DB-25 female",
		"DB-15 male",
		"DB-15 female",
		"DB-9 male",
		"DB-9 female",
		"RJ-11",
		"RJ-45",
		"50 Pin MiniSCSI",
		"Mini DIN",
		"Micro DIN",
		"PS/2",
		"Infrared",
		"HP-HIL",
		"Access Bus (USB)",
		"SSA SCSI",
		"Circular DIN-8 male",
		"Circular DIN-8 female",
		"On Board IDE",
		"On Board Floppy",
		"9 Pin Dual Inline (pin 10 cut)",
		"25 Pin Dual Inline (pin 26 cut)",
		"50 Pin Dual Inline",
		"68 Pin Dual Inline",
		"On Board Sound Input From CD-ROM",
		"Mini Centronics Type-14",
		"Mini Centronics Type-26",
		"Mini Jack (headphones)",
		"BNC",
		"IEEE 1394",
		"SAS/SATA Plug Receptacle" /* 0x22 */
	};
	static const char *type_0xA0[] = {
		"PC-98", /* 0xA0 */
		"PC-98 Hireso",
		"PC-H98",
		"PC-98 Note",
		"PC-98 Full" /* 0xA4 */
	};

	if (code <= 0x22)
		return type[code];
	if (code >= 0xA0 && code <= 0xA4)
		return type_0xA0[code - 0xA0];
	if (code == 0xFF)
		return "Other";
	return out_of_spec;
}

static const char *dmi_port_type(u8 code)
{
	/* 3.3.9.3 */
	static const char *type[] = {
		"None", /* 0x00 */
		"Parallel Port XT/AT Compatible",
		"Parallel Port PS/2",
		"Parallel Port ECP",
		"Parallel Port EPP",
		"Parallel Port ECP/EPP",
		"Serial Port XT/AT Compatible",
		"Serial Port 16450 Compatible",
		"Serial Port 16550 Compatible",
		"Serial Port 16550A Compatible",
		"SCSI Port",
		"MIDI Port",
		"Joystick Port",
		"Keyboard Port",
		"Mouse Port",
		"SSA SCSI",
		"USB",
		"Firewire (IEEE P1394)",
		"PCMCIA Type I",
		"PCMCIA Type II",
		"PCMCIA Type III",
		"Cardbus",
		"Access Bus Port",
		"SCSI II",
		"SCSI Wide",
		"PC-98",
		"PC-98 Hireso",
		"PC-H98",
		"Video Port",
		"Audio Port",
		"Modem Port",
		"Network Port",
		"SATA",
		"SAS" /* 0x21 */
	};
	static const char *type_0xA0[] = {
		"8251 Compatible", /* 0xA0 */
		"8251 FIFO Compatible" /* 0xA1 */
	};

	if (code <= 0x21)
		return type[code];
	if (code >= 0xA0 && code <= 0xA1)
		return type_0xA0[code - 0xA0];
	if (code == 0xFF)
		return "Other";
	return out_of_spec;
}

/*
 * 3.3.10 System Slots (Type 9)
 */

static const char *dmi_slot_type(u8 code)
{
	/* 3.3.10.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"ISA",
		"MCA",
		"EISA",
		"PCI",
		"PC Card (PCMCIA)",
		"VLB",
		"Proprietary",
		"Processor Card",
		"Proprietary Memory Card",
		"I/O Riser Card",
		"NuBus",
		"PCI-66",
		"AGP",
		"AGP 2x",
		"AGP 4x",
		"PCI-X",
		"AGP 8x" /* 0x13 */
	};
	static const char *type_0xA0[] = {
		"PC-98/C20", /* 0xA0 */
		"PC-98/C24",
		"PC-98/E",
		"PC-98/Local Bus",
		"PC-98/Card",
		"PCI Express",
		"PCI Express x1",
		"PCI Express x2",
		"PCI Express x4",
		"PCI Express x8",
		"PCI Express x16" /* 0xAA */
	};

	if (code >= 0x01 && code <= 0x13)
		return type[code - 0x01];
	if (code >= 0xA0 && code <= 0xAA)
		return type_0xA0[code - 0xA0];
	return out_of_spec;
}

static const char *dmi_slot_bus_width(u8 code)
{
	/* 3.3.10.2 */
	static const char *width[] = {
		"", /* 0x01, "Other" */
		"", /* "Unknown" */
		"8-bit ",
		"16-bit ",
		"32-bit ",
		"64-bit ",
		"128-bit ",
		"x1 ",
		"x2 ",
		"x4 ",
		"x8 ",
		"x12 ",
		"x16 ",
		"x32 " /* 0x0E */
	};

	if (code >= 0x01 && code <= 0x0E)
		return width[code - 0x01];
	return out_of_spec;
}

static const char *dmi_slot_current_usage(u8 code)
{
	/* 3.3.10.3 */
	static const char *usage[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Available",
		"In Use" /* 0x04 */
	};

	if (code >= 0x01 && code <= 0x04)
		return usage[code - 0x01];
	return out_of_spec;
}

static const char *dmi_slot_length(u8 code)
{
	/* 3.3.1O.4 */
	static const char *length[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Short",
		"Long" /* 0x04 */
	};

	if (code >= 0x01 && code <= 0x04)
		return length[code - 0x01];
	return out_of_spec;
}

static void dmi_slot_id(u8 code1, u8 code2, u8 type, const char *prefix)
{
	/* 3.3.10.5 */
	switch (type)
	{
		case 0x04: /* MCA */
			printf2("%sID: %u\n", prefix, code1);
			break;
		case 0x05: /* EISA */
			printf2("%sID: %u\n", prefix, code1);
			break;
		case 0x06: /* PCI */
		case 0x0E: /* PCI */
		case 0x0F: /* AGP */
		case 0x10: /* AGP */
		case 0x11: /* AGP */
		case 0x12: /* PCI-X */
		case 0x13: /* AGP */
		case 0xA5: /* PCI Express */
			printf2("%sID: %u\n", prefix, code1);
			break;
		case 0x07: /* PCMCIA */
			printf2("%sID: Adapter %u, Socket %u\n", prefix, code1, code2);
			break;
	}
}

static void dmi_slot_characteristics(u8 code1, u8 code2, const char *prefix)
{
	/* 3.3.10.6 */
	static const char *characteristics1[] = {
		"5.0 V is provided", /* 1 */
		"3.3 V is provided",
		"Opening is shared",
		"PC Card-16 is supported",
		"Cardbus is supported",
		"Zoom Video is supported",
		"Modem ring resume is supported" /* 7 */
	};
	/* 3.3.10.7 */
	static const char *characteristics2[] = {
		"PME signal is supported", /* 0 */
		"Hot-plug devices are supported",
		"SMBus signal is supported" /* 2 */
	};

	if (code1 & (1 << 0))
		printf2(" Unknown\n");
	else if ((code1 & 0xFE) == 0 && (code2 & 0x07) == 0)
		printf2(" None\n");
	else
	{
		int i;

		printf2("\n");
		for (i = 1; i <= 7; i++)
			if (code1 & (1 << i))
				printf2("%s%s\n", prefix, characteristics1[i - 1]);
		for (i = 0; i <= 2; i++)
			if (code2 & (1 << i))
				printf2("%s%s\n", prefix, characteristics2[i]);
	}
}

static void dmi_slot_segment_bus_func(u16 code1, u8 code2, u8 code3, const char *prefix)
{
	/* 3.3.10.8 */
	if (!(code1 == 0xFFFF && code2 == 0xFF && code3 == 0xFF))
		printf2("%sBus Address: %04x:%02x:%02x.%x\n",
		       prefix, code1, code2, code3 >> 3, code3 & 0x7);
}

/*
 * 3.3.11 On Board Devices Information (Type 10)
 */

static const char *dmi_on_board_devices_type(u8 code)
{
	/* 3.3.11.1 and 3.3.42.2 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Video",
		"SCSI Controller",
		"Ethernet",
		"Token Ring",
		"Sound",
		"PATA Controller",
		"SATA Controller",
		"SAS Controller" /* 0x0A */
	};

	if (code >= 0x01 && code <= 0x0A)
		return type[code - 0x01];
	return out_of_spec;
}

static void dmi_on_board_devices(const struct dmi_header *h, const char *prefix)
{
	u8 *p = h->data + 4;
	u8 count = (h->length - 0x04) / 2;
	int i;

	for (i = 0; i < count; i++)
	{
		if (count == 1)
			printf2("%sOn Board Device Information\n",
				prefix);
		else
			printf2("%sOn Board Device %d Information\n",
				prefix, i + 1);
		printf2("%s\tType: %s\n",
			prefix, dmi_on_board_devices_type(p[2 * i] & 0x7F));
		printf2("%s\tStatus: %s\n",
			prefix, p[2 * i] & 0x80 ? "Enabled" : "Disabled");
		printf2("%s\tDescription: %s\n",
			prefix, dmi_string(h, p[2 * i + 1]));
	}
}

/*
 * 3.3.12 OEM Strings (Type 11)
 */

static void dmi_oem_strings(const struct dmi_header *h, const char *prefix)
{
	u8 *p = h->data + 4;
	u8 count = p[0x00];
	int i;

	for (i = 1; i <= count; i++)
		printf2("%sString %d: %s\n",
			prefix, i, dmi_string(h, i));
}

/*
 * 3.3.13 System Configuration Options (Type 12)
 */

static void dmi_system_configuration_options(const struct dmi_header *h, const char *prefix)
{
	u8 *p = h->data + 4;
	u8 count = p[0x00];
	int i;

	for (i = 1; i <= count; i++)
		printf2("%sOption %d: %s\n",
			prefix, i, dmi_string(h, i));
}

/*
 * 3.3.14 BIOS Language Information (Type 13)
 */

static void dmi_bios_languages(const struct dmi_header *h, const char *prefix)
{
	u8 *p = h->data + 4;
	u8 count = p[0x00];
	int i;

	for (i = 1; i <= count; i++)
		printf2("%s%s\n",
			prefix, dmi_string(h, i));
}

/*
 * 3.3.15 Group Associations (Type 14)
 */

static void dmi_group_associations_items(u8 count, const u8 *p, const char *prefix)
{
	int i;

	for (i = 0; i < count; i++)
	{
		printf2("%s0x%04X (%s)\n",
			prefix, WORD(p + 3 * i + 1),
			dmi_smbios_structure_type(p[3 * i]));
	}
}

/*
 * 3.3.16 System Event Log (Type 15)
 */

static const char *dmi_event_log_method(u8 code)
{
	static const char *method[] = {
		"Indexed I/O, one 8-bit index port, one 8-bit data port", /* 0x00 */
		"Indexed I/O, two 8-bit index ports, one 8-bit data port",
		"Indexed I/O, one 16-bit index port, one 8-bit data port",
		"Memory-mapped physical 32-bit address",
		"General-purpose non-volatile data functions" /* 0x04 */
	};

	if (code <= 0x04)
		return method[code];
	if (code >= 0x80)
		return "OEM-specific";
	return out_of_spec;
}

static void dmi_event_log_status(u8 code)
{
	static const char *valid[] = {
		"Invalid", /* 0 */
		"Valid" /* 1 */
	};
	static const char *full[] = {
		"Not Full", /* 0 */
		"Full" /* 1 */
	};

	printf2(" %s, %s",
		valid[(code >> 0) & 1], full[(code >> 1) & 1]);
}

static void dmi_event_log_address(u8 method, const u8 *p)
{
	/* 3.3.16.3 */
	switch (method)
	{
		case 0x00:
		case 0x01:
		case 0x02:
			printf2(" Index 0x%04X, Data 0x%04X", WORD(p), WORD(p + 2));
			break;
		case 0x03:
			printf2(" 0x%08X", DWORD(p));
			break;
		case 0x04:
			printf2(" 0x%04X", WORD(p));
			break;
		default:
			printf2(" Unknown");
	}
}

static const char *dmi_event_log_header_type(u8 code)
{
	static const char *type[] = {
		"No Header", /* 0x00 */
		"Type 1" /* 0x01 */
	};

	if (code <= 0x01)
		return type[code];
	if (code >= 0x80)
		return "OEM-specific";
	return out_of_spec;
}

static const char *dmi_event_log_descriptor_type(u8 code)
{
	/* 3.3.16.6.1 */
	static const char *type[] = {
		NULL, /* 0x00 */
		"Single-bit ECC memory error",
		"Multi-bit ECC memory error",
		"Parity memory error",
		"Bus timeout",
		"I/O channel block",
		"Software NMI",
		"POST memory resize",
		"POST error",
		"PCI parity error",
		"PCI system error",
		"CPU failure",
		"EISA failsafe timer timeout",
		"Correctable memory log disabled",
		"Logging disabled",
		NULL, /* 0x0F */
		"System limit exceeded",
		"Asynchronous hardware timer expired",
		"System configuration information",
		"Hard disk information",
		"System reconfigured",
		"Uncorrectable CPU-complex error",
		"Log area reset/cleared",
		"System boot" /* 0x17 */
	};

	if (code <= 0x17 && type[code] != NULL)
		return type[code];
	if (code >= 0x80 && code <= 0xFE)
		return "OEM-specific";
	if (code == 0xFF)
		return "End of log";
	return out_of_spec;
}

static const char *dmi_event_log_descriptor_format(u8 code)
{
	/* 3.3.16.6.2 */
	static const char *format[] = {
		"None", /* 0x00 */
		"Handle",
		"Multiple-event",
		"Multiple-event handle",
		"POST results bitmap",
		"System management",
		"Multiple-event system management" /* 0x06 */
	};

	if (code <= 0x06)
		return format[code];
	if (code >= 0x80)
		return "OEM-specific";
	return out_of_spec;
}

static void dmi_event_log_descriptors(u8 count, u8 len, const u8 *p, const char *prefix)
{
	/* 3.3.16.1 */
	int i;

	for (i = 0; i < count; i++)
	{
		if (len >= 0x02)
		{
			printf2("%sDescriptor %u: %s\n",
				prefix, i + 1, dmi_event_log_descriptor_type(p[i * len]));
			printf2("%sData Format %u: %s\n",
				prefix, i + 1, dmi_event_log_descriptor_format(p[i * len + 1]));
		}
	}
}

/*
 * 3.3.17 Physical Memory Array (Type 16)
 */

static const char *dmi_memory_array_location(u8 code)
{
	/* 3.3.17.1 */
	static const char *location[] = {
		"Other", /* 0x01 */
		"Unknown",
		"System Board Or Motherboard",
		"ISA Add-on Card",
		"EISA Add-on Card",
		"PCI Add-on Card",
		"MCA Add-on Card",
		"PCMCIA Add-on Card",
		"Proprietary Add-on Card",
		"NuBus" /* 0x0A */
	};
	static const char *location_0xA0[] = {
		"PC-98/C20 Add-on Card", /* 0xA0 */
		"PC-98/C24 Add-on Card",
		"PC-98/E Add-on Card",
		"PC-98/Local Bus Add-on Card" /* 0xA3 */
	};

	if (code >= 0x01 && code <= 0x0A)
		return location[code - 0x01];
	if (code >= 0xA0 && code <= 0xA4)
		return location_0xA0[code - 0xA0];
	return out_of_spec;
}

static const char *dmi_memory_array_use(u8 code)
{
	/* 3.3.17.2 */
	static const char *use[] = {
		"Other", /* 0x01 */
		"Unknown",
		"System Memory",
		"Video Memory",
		"Flash Memory",
		"Non-volatile RAM",
		"Cache Memory" /* 0x07 */
	};

	if (code >= 0x01 && code <= 0x07)
		return use[code - 0x01];
	return out_of_spec;
}

static const char *dmi_memory_array_ec_type(u8 code)
{
	/* 3.3.17.3 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"None",
		"Parity",
		"Single-bit ECC",
		"Multi-bit ECC",
		"CRC" /* 0x07 */
	};

	if (code >= 0x01 && code <= 0x07)
		return type[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_array_capacity(u32 code)
{
	if (code == 0x8000000)
		printf2(" Unknown");
	else
	{
		if ((code & 0x000FFFFF) == 0)
			printf2(" %u GB", code >> 20);
		else if ((code & 0x000003FF) == 0)
			printf2(" %u MB", code >> 10);
		else
			printf2(" %u kB", code);
	}
}

static void dmi_memory_array_error_handle(u16 code)
{
	if (code == 0xFFFE)
		printf2(" Not Provided");
	else if (code == 0xFFFF)
		printf2(" No Error");
	else
		printf2(" 0x%04X", code);
}

/*
 * 3.3.18 Memory Device (Type 17)
 */

static void dmi_memory_device_width(u16 code)
{
	/*
	 * If no memory module is present, width may be 0
	 */
	if (code == 0xFFFF || code == 0)
		printf2(" Unknown");
	else
		printf2(" %u bits", code);
}

static void dmi_memory_device_size(u16 code)
{
	if (code == 0)
		printf2(" No Module Installed");
	else if (code == 0xFFFF)
		printf2(" Unknown");
	else
	{
		if (code & 0x8000)
			printf2(" %u kB", code & 0x7FFF);
		else
			printf2(" %u MB", code);
	}
}

static const char *dmi_memory_device_form_factor(u8 code)
{
	/* 3.3.18.1 */
	static const char *form_factor[] = {
		"Other", /* 0x01 */
		"Unknown",
		"SIMM",
		"SIP",
		"Chip",
		"DIP",
		"ZIP",
		"Proprietary Card",
		"DIMM",
		"TSOP",
		"Row Of Chips",
		"RIMM",
		"SODIMM",
		"SRIMM",
		"FB-DIMM" /* 0x0F */
	};

	if (code >= 0x01 && code <= 0x0F)
		return form_factor[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_device_set(u8 code)
{
	if (code == 0)
		printf2(" None");
	else if (code == 0xFF)
		printf2(" Unknown");
	else
		printf2(" %u", code);
}

static const char *dmi_memory_device_type(u8 code)
{
	/* 3.3.18.2 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"DRAM",
		"EDRAM",
		"VRAM",
		"SRAM",
		"RAM",
		"ROM",
		"Flash",
		"EEPROM",
		"FEPROM",
		"EPROM",
		"CDRAM",
		"3DRAM",
		"SDRAM",
		"SGRAM",
		"RDRAM",
		"DDR",
		"DDR2",
		"DDR2 FB-DIMM" /* 0x14 */
	};

	if (code >= 0x01 && code <= 0x14)
		return type[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_device_type_detail(u16 code)
{
	/* 3.3.18.3 */
	static const char *detail[] = {
		"Other", /* 1 */
		"Unknown",
		"Fast-paged",
		"Static Column",
		"Pseudo-static",
		"RAMBus",
		"Synchronous",
		"CMOS",
		"EDO",
		"Window DRAM",
		"Cache DRAM",
		"Non-Volatile" /* 12 */
	};

	if ((code & 0x1FFE) == 0)
		printf2(" None");
	else
	{
		int i;

		for (i = 1; i <= 12; i++)
			if (code & (1 << i))
				printf2(" %s", detail[i - 1]);
	}
}

static void dmi_memory_device_speed(u16 code)
{
	if (code == 0)
		printf2(" Unknown");
	else
		printf2(" %u MHz", code);
}

/*
 * 3.3.19 32-bit Memory Error Information (Type 18)
 */

static const char *dmi_memory_error_type(u8 code)
{
	/* 3.3.19.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"OK",
		"Bad Read",
		"Parity Error",
		"Single-bit Error",
		"Double-bit Error",
		"Multi-bit Error",
		"Nibble Error",
		"Checksum Error",
		"CRC Error",
		"Corrected Single-bit Error",
		"Corrected Error",
		"Uncorrectable Error" /* 0x0E */
	};

	if (code >= 0x01 && code <= 0x0E)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_memory_error_granularity(u8 code)
{
	/* 3.3.19.2 */
	static const char *granularity[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Device Level",
		"Memory Partition Level" /* 0x04 */
	};

	if (code >= 0x01 && code <= 0x04)
		return granularity[code - 0x01];
	return out_of_spec;
}

static const char *dmi_memory_error_operation(u8 code)
{
	/* 3.3.19.3 */
	static const char *operation[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Read",
		"Write",
		"Partial Write" /* 0x05 */
	};

	if (code >= 0x01 && code <= 0x05)
		return operation[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_error_syndrome(u32 code)
{
	if (code == 0x00000000)
		printf2(" Unknown");
	else
		printf2(" 0x%08X", code);
}

static void dmi_32bit_memory_error_address(u32 code)
{
	if (code == 0x80000000)
		printf2(" Unknown");
	else
		printf2(" 0x%08X", code);
}

/*
 * 3.3.20 Memory Array Mapped Address (Type 19)
 */

static void dmi_mapped_address_size(u32 code)
{
	if (code == 0)
		printf2(" Invalid");
	else if ((code & 0x000FFFFF) == 0)
		printf2(" %u GB", code >> 20);
	else if ((code & 0x000003FF) == 0)
		printf2(" %u MB", code >> 10);
	else
		printf2(" %u kB", code);
}

/*
 * 3.3.21 Memory Device Mapped Address (Type 20)
 */

static void dmi_mapped_address_row_position(u8 code)
{
	if (code == 0)
		printf2(" %s", out_of_spec);
	else if (code == 0xFF)
		printf2(" Unknown");
	else
		printf2(" %u", code);
}

static void dmi_mapped_address_interleave_position(u8 code, const char *prefix)
{
	if (code != 0)
	{
		printf2("%sInterleave Position:", prefix);
		if (code == 0xFF)
			printf2(" Unknown");
		else
			printf2(" %u", code);
		printf2("\n");
	}
}

static void dmi_mapped_address_interleaved_data_depth(u8 code, const char *prefix)
{
	if (code != 0)
	{
		printf2("%sInterleaved Data Depth:", prefix);
		if (code == 0xFF)
			printf2(" Unknown");
		else
			printf2(" %u", code);
		printf2("\n");
	}
}

/*
 * 3.3.22 Built-in Pointing Device (Type 21)
 */

static const char *dmi_pointing_device_type(u8 code)
{
	/* 3.3.22.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Mouse",
		"Track Ball",
		"Track Point",
		"Glide Point",
		"Touch Pad",
		"Touch Screen",
		"Optical Sensor" /* 0x09 */
	};

	if (code >= 0x01 && code <= 0x09)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_pointing_device_interface(u8 code)
{
	/* 3.3.22.2 */
	static const char *interface[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Serial",
		"PS/2",
		"Infrared",
		"HIP-HIL",
		"Bus Mouse",
		"ADB (Apple Desktop Bus)" /* 0x08 */
	};
	static const char *interface_0xA0[] = {
		"Bus Mouse DB-9", /* 0xA0 */
		"Bus Mouse Micro DIN",
		"USB" /* 0xA2 */
	};

	if (code >= 0x01 && code <= 0x08)
		return interface[code - 0x01];
	if (code >= 0xA0 && code <= 0xA2)
		return interface_0xA0[code - 0xA0];
	return out_of_spec;
}

/*
 * 3.3.23 Portable Battery (Type 22)
 */

static const char *dmi_battery_chemistry(u8 code)
{
	/* 3.3.23.1 */
	static const char *chemistry[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Lead Acid",
		"Nickel Cadmium",
		"Nickel Metal Hydride",
		"Lithium Ion",
		"Zinc Air",
		"Lithium Polymer" /* 0x08 */
	};

	if (code >= 0x01 && code <= 0x08)
		return chemistry[code - 0x01];
	return out_of_spec;
}

static void dmi_battery_capacity(u16 code, u8 multiplier)
{
	if (code == 0)
		printf2(" Unknown");
	else
		printf2(" %u mWh", code * multiplier);
}

static void dmi_battery_voltage(u16 code)
{
	if (code == 0)
		printf2(" Unknown");
	else
		printf2(" %u mV", code);
}

static void dmi_battery_maximum_error(u8 code)
{
	if (code == 0xFF)
		printf2(" Unknown");
	else
		printf2(" %u%%", code);
}

/*
 * 3.3.24 System Reset (Type 23)
 */

/* code is assumed to be a 2-bit value */
static const char *dmi_system_reset_boot_option(u8 code)
{
	static const char *option[] = {
		out_of_spec, /* 0x0 */
		"Operating System", /* 0x1 */
		"System Utilities",
		"Do Not Reboot" /* 0x3 */
	};

	return option[code];
}

static void dmi_system_reset_count(u16 code)
{
	if (code == 0xFFFF)
		printf2(" Unknown");
	else
		printf2(" %u", code);
}

static void dmi_system_reset_timer(u16 code)
{
	if (code == 0xFFFF)
		printf2(" Unknown");
	else
		printf2(" %u min", code);
}

/*
 * 3.3.25 Hardware Security (Type 24)
 */

static const char *dmi_hardware_security_status(u8 code)
{
	static const char *status[] = {
		"Disabled", /* 0x00 */
		"Enabled",
		"Not Implemented",
		"Unknown" /* 0x03 */
	};

	return status[code];
}

/*
 * 3.3.26 System Power Controls (Type 25)
 */

static void dmi_power_controls_power_on(const u8 *p)
{
	/* 3.3.26.1 */
	if (dmi_bcd_range(p[0], 0x01, 0x12))
		printf2(" %02X", p[0]);
	else
		printf2(" *");
	if (dmi_bcd_range(p[1], 0x01, 0x31))
		printf2("-%02X", p[1]);
	else
		printf2("-*");
	if (dmi_bcd_range(p[2], 0x00, 0x23))
		printf2(" %02X", p[2]);
	else
		printf2(" *");
	if (dmi_bcd_range(p[3], 0x00, 0x59))
		printf2(":%02X", p[3]);
	else
		printf2(":*");
	if (dmi_bcd_range(p[4], 0x00, 0x59))
		printf2(":%02X", p[4]);
	else
		printf2(":*");
}

/*
 * 3.3.27 Voltage Probe (Type 26)
 */

static const char *dmi_voltage_probe_location(u8 code)
{
	/* 3.3.27.1 */
	static const char *location[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Processor",
		"Disk",
		"Peripheral Bay",
		"System Management Module",
		"Motherboard",
		"Memory Module",
		"Processor Module",
		"Power Unit",
		"Add-in Card" /* 0x0B */
	};

	if (code >= 0x01 && code <= 0x0B)
		return location[code - 0x01];
	return out_of_spec;
}

static const char *dmi_probe_status(u8 code)
{
	/* 3.3.27.1 */
	static const char *status[] = {
		"Other", /* 0x01 */
		"Unknown",
		"OK",
		"Non-critical",
		"Critical",
		"Non-recoverable" /* 0x06 */
	};

	if (code >= 0x01 && code <= 0x06)
		return status[code - 0x01];
	return out_of_spec;
}

static void dmi_voltage_probe_value(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.3f V", (float)(i16)code / 1000);
}

static void dmi_voltage_probe_resolution(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.1f mV", (float)code / 10);
}

static void dmi_probe_accuracy(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.2f%%", (float)code / 100);
}

/*
 * 3.3.28 Cooling Device (Type 27)
 */

static const char *dmi_cooling_device_type(u8 code)
{
	/* 3.3.28.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Fan",
		"Centrifugal Blower",
		"Chip Fan",
		"Cabinet Fan",
		"Power Supply Fan",
		"Heat Pipe",
		"Integrated Refrigeration" /* 0x09 */
	};
	static const char *type_0x10[] = {
		"Active Cooling", /* 0x10 */
		"Passive Cooling" /* 0x11 */
	};

	if (code >= 0x01 && code <= 0x09)
		return type[code - 0x01];
	if (code >= 0x10 && code <= 0x11)
		return type_0x10[code - 0x10];
	return out_of_spec;
}

static void dmi_cooling_device_speed(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown Or Non-rotating");
	else
		printf2(" %u rpm", code);
}

/*
 * 3.3.29 Temperature Probe (Type 28)
 */

static const char *dmi_temperature_probe_location(u8 code)
{
	/* 3.3.29.1 */
	static const char *location[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Processor",
		"Disk",
		"Peripheral Bay",
		"System Management Module",
		"Motherboard",
		"Memory Module",
		"Processor Module",
		"Power Unit",
		"Add-in Card",
		"Front Panel Board",
		"Back Panel Board",
		"Power System Board",
		"Drive Back Plane" /* 0x0F */
	};

	if (code >= 0x01 && code <= 0x0F)
		return location[code - 0x01];
	return out_of_spec;
}

static void dmi_temperature_probe_value(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.1f deg C", (float)(i16)code / 10);
}

static void dmi_temperature_probe_resolution(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.3f deg C", (float)code / 1000);
}

/*
 * 3.3.30 Electrical Current Probe (Type 29)
 */

static void dmi_current_probe_value(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.3f A", (float)(i16)code / 1000);
}

static void dmi_current_probe_resolution(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.1f mA", (float)code / 10);
}

/*
 * 3.3.33 System Boot Information (Type 32)
 */

static const char *dmi_system_boot_status(u8 code)
{
	static const char *status[] = {
		"No errors detected", /* 0 */
		"No bootable media",
		"Operating system failed to load",
		"Firmware-detected hardware failure",
		"Operating system-detected hardware failure",
		"User-requested boot",
		"System security violation",
		"Previously-requested image",
		"System watchdog timer expired" /* 8 */
	};

	if (code <= 8)
		return status[code];
	if (code >= 128 && code <= 191)
		return "OEM-specific";
	if (code >= 192)
		return "Product-specific";
	return out_of_spec;
}

/*
 * 3.3.34 64-bit Memory Error Information (Type 33)
 */

static void dmi_64bit_memory_error_address(u64 code)
{
	if (code.h == 0x80000000 && code.l == 0x00000000)
		printf2(" Unknown");
	else
		printf2(" 0x%08X%08X", code.h, code.l);
}

/*
 * 3.3.35 Management Device (Type 34)
 */

static const char *dmi_management_device_type(u8 code)
{
	/* 3.3.35.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"LM75",
		"LM78",
		"LM79",
		"LM80",
		"LM81",
		"ADM9240",
		"DS1780",
		"MAX1617",
		"GL518SM",
		"W83781D",
		"HT82H791" /* 0x0D */
	};

	if (code >= 0x01 && code <= 0x0D)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_management_device_address_type(u8 code)
{
	/* 3.3.35.2 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"I/O Port",
		"Memory",
		"SMBus" /* 0x05 */
	};

	if (code >= 0x01 && code <= 0x05)
		return type[code - 0x01];
	return out_of_spec;
}

/*
 * 3.3.38 Memory Channel (Type 37)
 */

static const char *dmi_memory_channel_type(u8 code)
{
	/* 3.3.38.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"RamBus",
		"SyncLink" /* 0x04 */
	};

	if (code >= 0x01 && code <= 0x04)
		return type[code - 0x01];
	return out_of_spec;
}

static void dmi_memory_channel_devices(u8 count, const u8 *p, const char *prefix)
{
	int i;

	for (i = 1; i <= count; i++)
	{
		printf2("%sDevice %u Load: %u\n",
			prefix, i, p[3 * i]);
		if (!(opt.flags & FLAG_QUIET))
			printf2("%sDevice %u Handle: 0x%04X\n",
				prefix, i, WORD(p + 3 * i + 1));
	}
}

/*
 * 3.3.39 IPMI Device Information (Type 38)
 */

static const char *dmi_ipmi_interface_type(u8 code)
{
	/* 3.3.39.1 and IPMI 2.0, appendix C1, table C1-2 */
	static const char *type[] = {
		"Unknown", /* 0x00 */
		"KCS (Keyboard Control Style)",
		"SMIC (Server Management Interface Chip)",
		"BT (Block Transfer)",
		"SSIF (SMBus System Interface)" /* 0x04 */
	};

	if (code <= 0x04)
		return type[code];
	return out_of_spec;
}

static void dmi_ipmi_base_address(u8 type, const u8 *p, u8 lsb)
{
	if (type == 0x04) /* SSIF */
	{
		printf2("0x%02X (SMBus)", (*p) >> 1);
	}
	else
	{
		u64 address = QWORD(p);
		printf2("0x%08X%08X (%s)", address.h, (address.l & ~1) | lsb,
			address.l & 1 ? "I/O" : "Memory-mapped");
	}
}

/* code is assumed to be a 2-bit value */
static const char *dmi_ipmi_register_spacing(u8 code)
{
	/* IPMI 2.0, appendix C1, table C1-1 */
	static const char *spacing[] = {
		"Successive Byte Boundaries", /* 0x00 */
		"32-bit Boundaries",
		"16-byte Boundaries", /* 0x02 */
		out_of_spec /* 0x03 */
	};

	return spacing[code];
}

/*
 * 3.3.40 System Power Supply (Type 39)
 */

static void dmi_power_supply_power(u16 code)
{
	if (code == 0x8000)
		printf2(" Unknown");
	else
		printf2(" %.3f W", (float)code / 1000);
}

static const char *dmi_power_supply_type(u8 code)
{
	/* 3.3.40.1 */
	static const char *type[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Linear",
		"Switching",
		"Battery",
		"UPS",
		"Converter",
		"Regulator" /* 0x08 */
	};

	if (code >= 0x01 && code <= 0x08)
		return type[code - 0x01];
	return out_of_spec;
}

static const char *dmi_power_supply_status(u8 code)
{
	/* 3.3.40.1 */
	static const char *status[] = {
		"Other", /* 0x01 */
		"Unknown",
		"OK",
		"Non-critical",
		"Critical" /* 0x05 */
	};

	if (code >= 0x01 && code <= 0x05)
		return status[code - 0x01];
	return out_of_spec;
}

static const char *dmi_power_supply_range_switching(u8 code)
{
	/* 3.3.40.1 */
	static const char *switching[] = {
		"Other", /* 0x01 */
		"Unknown",
		"Manual",
		"Auto-switch",
		"Wide Range",
		"N/A" /* 0x06 */
	};

	if (code >= 0x01 && code <= 0x06)
		return switching[code - 0x01];
	return out_of_spec;
}

/*
 * 3.3.41 Additional Information (Type 40)
 *
 * Proper support of this entry type would require redesigning a large part of
 * the code, so I am waiting to see actual implementations of it to decide
 * whether it's worth the effort.
 */

static void dmi_additional_info(const struct dmi_header *h, const char *prefix)
{
	u8 *p = h->data + 4;
	u8 count = *p++;
	u8 length;
	int i, offset = 5;

	for (i = 0; i < count; i++)
	{
		printf2("%sAdditional Information %d\n", prefix, i + 1);

		/* Check for short entries */
		if (h->length < offset + 1) break;
		length = p[0x00];
		if (length < 0x05 || h->length < offset + length) break;

		printf2("%s\tReferenced Handle: 0x%04x\n",
			prefix, WORD(p + 0x01));
		printf2("%s\tReferenced Offset: 0x%02x\n",
			prefix, p[0x03]);
		printf2("%s\tString: %s\n",
			prefix, dmi_string(h, p[0x04]));

		printf2("%s\tValue: ", prefix);
		switch (length - 0x05)
		{
			case 1:
				printf2("0x%02x", p[0x05]);
				break;
			case 2:
				printf2("0x%04x", WORD(p + 0x05));
				break;
			case 4:
				printf2("0x%08x", DWORD(p + 0x05));
				break;
			default:
				printf2("Unexpected size");
				break;
		}
		printf2("\n");

		p += length;
		offset += length;
	}
}

/*
 * Main
 */

static void dmi_decode(const struct dmi_header *h, u16 ver)
{
	const u8 *data = h->data;

	/*
	 * Note: DMI types 37, 39 and 40 are untested
	 */
	switch (h->type)
	{
		case 0: /* 3.3.1 BIOS Information */
			printf2("BIOS Information\n");
			if (h->length < 0x12) break;
			printf2("\tVendor: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tVersion: %s\n",
				dmi_string(h, data[0x05]));
			printf2("\tRelease Date: %s\n",
				dmi_string(h, data[0x08]));
			/*
			 * On IA-64, the BIOS base address will read 0 because
			 * there is no BIOS. Skip the base address and the
			 * runtime size in this case.
			 */
			if (WORD(data + 0x06) != 0)
			{
				printf2("\tAddress: 0x%04X0\n",
					WORD(data + 0x06));
				printf2("\tRuntime Size:");
				dmi_bios_runtime_size((0x10000 - WORD(data + 0x06)) << 4);
				printf2("\n");
			}
			printf2("\tROM Size: %u kB\n",
				(data[0x09] + 1) << 6);
			printf2("\tCharacteristics:\n");
			dmi_bios_characteristics(QWORD(data + 0x0A), "\t\t");
			if (h->length < 0x13) break;
			dmi_bios_characteristics_x1(data[0x12], "\t\t");
			if (h->length < 0x14) break;
			dmi_bios_characteristics_x2(data[0x13], "\t\t");
			if (h->length < 0x18) break;
			if (data[0x14] != 0xFF && data[0x15] != 0xFF)
				printf2("\tBIOS Revision: %u.%u\n",
					data[0x14], data[0x15]);
			if (data[0x16] != 0xFF && data[0x17] != 0xFF)
				printf2("\tFirmware Revision: %u.%u\n",
					data[0x16], data[0x17]);
			break;

		case 1: /* 3.3.2 System Information */
			printf2("System Information\n");
			if (h->length < 0x08) break;
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tProduct Name: %s\n",
				dmi_string(h, data[0x05]));
			printf2("\tVersion: %s\n",
				dmi_string(h, data[0x06]));
			printf2("\tSerial Number: %s\n",
				dmi_string(h, data[0x07]));
			if (h->length < 0x19) break;
			printf2("\tUUID: ");
			dmi_system_uuid(data + 0x08, ver);
			printf2("\n");
			printf2("\tWake-up Type: %s\n",
				dmi_system_wake_up_type(data[0x18]));
			if (h->length < 0x1B) break;
			printf2("\tSKU Number: %s\n",
				dmi_string(h, data[0x19]));
			printf2("\tFamily: %s\n",
				dmi_string(h, data[0x1A]));
			break;

		case 2: /* 3.3.3 Base Board Information */
			printf2("Base Board Information\n");
			if (h->length < 0x08) break;
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tProduct Name: %s\n",
				dmi_string(h, data[0x05]));
			printf2("\tVersion: %s\n",
				dmi_string(h, data[0x06]));
			printf2("\tSerial Number: %s\n",
				dmi_string(h, data[0x07]));
			if (h->length < 0x09) break;
			printf2("\tAsset Tag: %s\n",
				dmi_string(h, data[0x08]));
			if (h->length < 0x0A) break;
			printf2("\tFeatures:");
			dmi_base_board_features(data[0x09], "\t\t");
			if (h->length < 0x0F) break;
			printf2("\tLocation In Chassis: %s\n",
				dmi_string(h, data[0x0A]));
			if (!(opt.flags & FLAG_QUIET))
				printf2("\tChassis Handle: 0x%04X\n",
					WORD(data + 0x0B));
			printf2("\tType: %s\n",
				dmi_base_board_type(data[0x0D]));
			if (h->length < 0x0F + data[0x0E] * sizeof(u16)) break;
			if (!(opt.flags & FLAG_QUIET))
				dmi_base_board_handles(data[0x0E], data + 0x0F, "\t");
			break;

		case 3: /* 3.3.4 Chassis Information */
			printf2("Chassis Information\n");
			if (h->length < 0x09) break;
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tType: %s\n",
				dmi_chassis_type(data[0x05] & 0x7F));
			printf2("\tLock: %s\n",
				dmi_chassis_lock(data[0x05] >> 7));
			printf2("\tVersion: %s\n",
				dmi_string(h, data[0x06]));
			printf2("\tSerial Number: %s\n",
				dmi_string(h, data[0x07]));
			printf2("\tAsset Tag: %s\n",
				dmi_string(h, data[0x08]));
			if (h->length < 0x0D) break;
			printf2("\tBoot-up State: %s\n",
				dmi_chassis_state(data[0x09]));
			printf2("\tPower Supply State: %s\n",
				dmi_chassis_state(data[0x0A]));
			printf2("\tThermal State: %s\n",
				dmi_chassis_state(data[0x0B]));
			printf2("\tSecurity Status: %s\n",
				dmi_chassis_security_status(data[0x0C]));
			if (h->length < 0x11) break;
			printf2("\tOEM Information: 0x%08X\n",
				DWORD(data + 0x0D));
			if (h->length < 0x13) break;
			printf2("\tHeight:");
			dmi_chassis_height(data[0x11]);
			printf2("\n");
			printf2("\tNumber Of Power Cords:");
			dmi_chassis_power_cords(data[0x12]);
			printf2("\n");
			if (h->length < 0x15) break;
			if (h->length < 0x15 + data[0x13] * data[0x14]) break;
			dmi_chassis_elements(data[0x13], data[0x14], data + 0x15, "\t");
			break;

		case 4: /* 3.3.5 Processor Information */
			printf2("Processor Information\n");
			if (h->length < 0x1A) break;
			printf2("\tSocket Designation: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tType: %s\n",
				dmi_processor_type(data[0x05]));
			printf2("\tFamily: %s\n",
				dmi_processor_family(h));
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x07]));
			dmi_processor_id(data[0x06], data + 8, dmi_string(h, data[0x10]), "\t");
			printf2("\tVersion: %s\n",
				dmi_string(h, data[0x10]));
			printf2("\tVoltage:");
			dmi_processor_voltage(data[0x11]);
			printf2("\n");
			printf2("\tExternal Clock: ");
			dmi_processor_frequency(data + 0x12);
			printf2("\n");
			printf2("\tMax Speed: ");
			dmi_processor_frequency(data + 0x14);
			printf2("\n");
			printf2("\tCurrent Speed: ");
			dmi_processor_frequency(data + 0x16);
			printf2("\n");
			if (data[0x18] & (1 << 6))
				printf2("\tStatus: Populated, %s\n",
					dmi_processor_status(data[0x18] & 0x07));
			else
				printf2("\tStatus: Unpopulated\n");
			printf2("\tUpgrade: %s\n",
				dmi_processor_upgrade(data[0x19]));
			if (h->length < 0x20) break;
			if (!(opt.flags & FLAG_QUIET))
			{
				printf2("\tL1 Cache Handle:");
				dmi_processor_cache(WORD(data + 0x1A), "L1", ver);
				printf2("\n");
				printf2("\tL2 Cache Handle:");
				dmi_processor_cache(WORD(data + 0x1C), "L2", ver);
				printf2("\n");
				printf2("\tL3 Cache Handle:");
				dmi_processor_cache(WORD(data + 0x1E), "L3", ver);
				printf2("\n");
			}
			if (h->length < 0x23) break;
			printf2("\tSerial Number: %s\n",
				dmi_string(h, data[0x20]));
			printf2("\tAsset Tag: %s\n",
				dmi_string(h, data[0x21]));
			printf2("\tPart Number: %s\n",
				dmi_string(h, data[0x22]));
			if (h->length < 0x28) break;
			if (data[0x23] != 0)
				printf2("\tCore Count: %u\n", data[0x23]);
			if (data[0x24] != 0)
				printf2("\tCore Enabled: %u\n", data[0x24]);
			if (data[0x25] != 0)
				printf2("\tThread Count: %u\n", data[0x25]);
			printf2("\tCharacteristics:");
			dmi_processor_characteristics(WORD(data + 0x26), "\t\t");
			break;

		case 5: /* 3.3.6 Memory Controller Information */
			printf2("Memory Controller Information\n");
			if (h->length < 0x0F) break;
			printf2("\tError Detecting Method: %s\n",
				dmi_memory_controller_ed_method(data[0x04]));
			printf2("\tError Correcting Capabilities:");
			dmi_memory_controller_ec_capabilities(data[0x05], "\t\t");
			printf2("\tSupported Interleave: %s\n",
				dmi_memory_controller_interleave(data[0x06]));
			printf2("\tCurrent Interleave: %s\n",
				dmi_memory_controller_interleave(data[0x07]));
			printf2("\tMaximum Memory Module Size: %u MB\n",
				1 << data[0x08]);
			printf2("\tMaximum Total Memory Size: %u MB\n",
				data[0x0E] * (1 << data[0x08]));
			printf2("\tSupported Speeds:");
			dmi_memory_controller_speeds(WORD(data + 0x09), "\t\t");
			printf2("\tSupported Memory Types:");
			dmi_memory_module_types(WORD(data + 0x0B), "\n\t\t");
			printf2("\n");
			printf2("\tMemory Module Voltage:");
			dmi_processor_voltage(data[0x0D]);
			printf2("\n");
			if (h->length < 0x0F + data[0x0E] * sizeof(u16)) break;
			dmi_memory_controller_slots(data[0x0E], data + 0x0F, "\t");
			if (h->length < 0x10 + data[0x0E] * sizeof(u16)) break;
			printf2("\tEnabled Error Correcting Capabilities:");
			dmi_memory_controller_ec_capabilities(data[0x0F + data[0x0E] * sizeof(u16)], "\t\t");
			break;

		case 6: /* 3.3.7 Memory Module Information */
			printf2("Memory Module Information\n");
			if (h->length < 0x0C) break;
			printf2("\tSocket Designation: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tBank Connections:");
			dmi_memory_module_connections(data[0x05]);
			printf2("\n");
			printf2("\tCurrent Speed:");
			dmi_memory_module_speed(data[0x06]);
			printf2("\n");
			printf2("\tType:");
			dmi_memory_module_types(WORD(data + 0x07), " ");
			printf2("\n");
			printf2("\tInstalled Size:");
			dmi_memory_module_size(data[0x09]);
			printf2("\n");
			printf2("\tEnabled Size:");
			dmi_memory_module_size(data[0x0A]);
			printf2("\n");
			printf2("\tError Status:");
			dmi_memory_module_error(data[0x0B], "\t\t");
			break;

		case 7: /* 3.3.8 Cache Information */
			printf2("Cache Information\n");
			if (h->length < 0x0F) break;
			printf2("\tSocket Designation: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tConfiguration: %s, %s, Level %u\n",
				WORD(data + 0x05) & 0x0080 ? "Enabled" : "Disabled",
				WORD(data + 0x05) & 0x0008 ? "Socketed" : "Not Socketed",
				(WORD(data + 0x05) & 0x0007) + 1);
			printf2("\tOperational Mode: %s\n",
				dmi_cache_mode((WORD(data + 0x05) >> 8) & 0x0003));
			printf2("\tLocation: %s\n",
				dmi_cache_location((WORD(data + 0x05) >> 5) & 0x0003));
			printf2("\tInstalled Size:");
			dmi_cache_size(WORD(data + 0x09));
			printf2("\n");
			printf2("\tMaximum Size:");
			dmi_cache_size(WORD(data + 0x07));
			printf2("\n");
			printf2("\tSupported SRAM Types:");
			dmi_cache_types(WORD(data + 0x0B), "\n\t\t");
			printf2("\n");
			printf2("\tInstalled SRAM Type:");
			dmi_cache_types(WORD(data + 0x0D), " ");
			printf2("\n");
			if (h->length < 0x13) break;
			printf2("\tSpeed:");
			dmi_memory_module_speed(data[0x0F]);
			printf2("\n");
			printf2("\tError Correction Type: %s\n",
				dmi_cache_ec_type(data[0x10]));
			printf2("\tSystem Type: %s\n",
				dmi_cache_type(data[0x11]));
			printf2("\tAssociativity: %s\n",
				dmi_cache_associativity(data[0x12]));
			break;

		case 8: /* 3.3.9 Port Connector Information */
			printf2("Port Connector Information\n");
			if (h->length < 0x09) break;
			printf2("\tInternal Reference Designator: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tInternal Connector Type: %s\n",
				dmi_port_connector_type(data[0x05]));
			printf2("\tExternal Reference Designator: %s\n",
				dmi_string(h, data[0x06]));
			printf2("\tExternal Connector Type: %s\n",
				dmi_port_connector_type(data[0x07]));
			printf2("\tPort Type: %s\n",
				dmi_port_type(data[0x08]));
			break;

		case 9: /* 3.3.10 System Slots */
			printf2("System Slot Information\n");
			if (h->length < 0x0C) break;
			printf2("\tDesignation: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tType: %s%s\n",
				dmi_slot_bus_width(data[0x06]),
				dmi_slot_type(data[0x05]));
			printf2("\tCurrent Usage: %s\n",
				dmi_slot_current_usage(data[0x07]));
			printf2("\tLength: %s\n",
				dmi_slot_length(data[0x08]));
			dmi_slot_id(data[0x09], data[0x0A], data[0x05], "\t");
			printf2("\tCharacteristics:");
			if (h->length < 0x0D)
				dmi_slot_characteristics(data[0x0B], 0x00, "\t\t");
			else
				dmi_slot_characteristics(data[0x0B], data[0x0C], "\t\t");
			if (h->length < 0x11) break;
			dmi_slot_segment_bus_func(WORD(data + 0x0D), data[0x0F], data[0x10], "\t");
			break;

		case 10: /* 3.3.11 On Board Devices Information */
			dmi_on_board_devices(h, "");
			break;

		case 11: /* 3.3.12 OEM Strings */
			printf2("OEM Strings\n");
			if (h->length < 0x05) break;
			dmi_oem_strings(h, "\t");
			break;

		case 12: /* 3.3.13 System Configuration Options */
			printf2("System Configuration Options\n");
			if (h->length < 0x05) break;
			dmi_system_configuration_options(h, "\t");
			break;

		case 13: /* 3.3.14 BIOS Language Information */
			printf2("BIOS Language Information\n");
			if (h->length < 0x16) break;
			printf2("\tInstallable Languages: %u\n", data[0x04]);
			dmi_bios_languages(h, "\t\t");
			printf2("\tCurrently Installed Language: %s\n",
				dmi_string(h, data[0x15]));
			break;

		case 14: /* 3.3.15 Group Associations */
			printf2("Group Associations\n");
			if (h->length < 0x05) break;
			printf2("\tName: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tItems: %u\n",
				(h->length - 0x05) / 3);
			dmi_group_associations_items((h->length - 0x05) / 3, data + 0x05, "\t\t");
			break;

		case 15: /* 3.3.16 System Event Log */
			printf2("System Event Log\n");
			if (h->length < 0x14) break;
			printf2("\tArea Length: %u bytes\n",
				WORD(data + 0x04));
			printf2("\tHeader Start Offset: 0x%04X\n",
				WORD(data + 0x06));
			if (WORD(data + 0x08) - WORD(data + 0x06))
				printf2("\tHeader Length: %u byte%s\n",
					WORD(data + 0x08) - WORD(data + 0x06),
					WORD(data + 0x08) - WORD(data + 0x06) > 1 ? "s" : "");
			printf2("\tData Start Offset: 0x%04X\n",
				WORD(data + 0x08));
			printf2("\tAccess Method: %s\n",
				dmi_event_log_method(data[0x0A]));
			printf2("\tAccess Address:");
			dmi_event_log_address(data[0x0A], data + 0x10);
			printf2("\n");
			printf2("\tStatus:");
			dmi_event_log_status(data[0x0B]);
			printf2("\n");
			printf2("\tChange Token: 0x%08X\n",
				DWORD(data + 0x0C));
			if (h->length < 0x17) break;
			printf2("\tHeader Format: %s\n",
				dmi_event_log_header_type(data[0x14]));
			printf2("\tSupported Log Type Descriptors: %u\n",
				data[0x15]);
			if (h->length < 0x17 + data[0x15] * data[0x16]) break;
			dmi_event_log_descriptors(data[0x15], data[0x16], data + 0x17, "\t");
			break;

		case 16: /* 3.3.17 Physical Memory Array */
			printf2("Physical Memory Array\n");
			if (h->length < 0x0F) break;
			printf2("\tLocation: %s\n",
				dmi_memory_array_location(data[0x04]));
			printf2("\tUse: %s\n",
				dmi_memory_array_use(data[0x05]));
			printf2("\tError Correction Type: %s\n",
				dmi_memory_array_ec_type(data[0x06]));
			printf2("\tMaximum Capacity:");
			dmi_memory_array_capacity(DWORD(data + 0x07));
			printf2("\n");
			if (!(opt.flags & FLAG_QUIET))
			{
				printf2("\tError Information Handle:");
				dmi_memory_array_error_handle(WORD(data + 0x0B));
				printf2("\n");
			}
			printf2("\tNumber Of Devices: %u\n",
				WORD(data + 0x0D));
			break;

		case 17: /* 3.3.18 Memory Device */
			printf2("Memory Device\n");
			if (h->length < 0x15) break;
			if (!(opt.flags & FLAG_QUIET))
			{
				printf2("\tArray Handle: 0x%04X\n",
					WORD(data + 0x04));
				printf2("\tError Information Handle:");
				dmi_memory_array_error_handle(WORD(data + 0x06));
				printf2("\n");
			}
			printf2("\tTotal Width:");
			dmi_memory_device_width(WORD(data + 0x08));
			printf2("\n");
			printf2("\tData Width:");
			dmi_memory_device_width(WORD(data + 0x0A));
			printf2("\n");
			printf2("\tSize:");
			dmi_memory_device_size(WORD(data + 0x0C));
			printf2("\n");
			printf2("\tForm Factor: %s\n",
				dmi_memory_device_form_factor(data[0x0E]));
			printf2("\tSet:");
			dmi_memory_device_set(data[0x0F]);
			printf2("\n");
			printf2("\tLocator: %s\n",
				dmi_string(h, data[0x10]));
			printf2("\tBank Locator: %s\n",
				dmi_string(h, data[0x11]));
			printf2("\tType: %s\n",
				dmi_memory_device_type(data[0x12]));
			printf2("\tType Detail:");
			dmi_memory_device_type_detail(WORD(data + 0x13));
			printf2("\n");
			if (h->length < 0x17) break;
			printf2("\tSpeed:");
			dmi_memory_device_speed(WORD(data + 0x15));
			printf2("\n");
			if (h->length < 0x1B) break;
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x17]));
			printf2("\tSerial Number: %s\n",
				dmi_string(h, data[0x18]));
			printf2("\tAsset Tag: %s\n",
				dmi_string(h, data[0x19]));
			printf2("\tPart Number: %s\n",
				dmi_string(h, data[0x1A]));
			if (h->length < 0x1C) break;
			printf2("\tRank: ");
			if ((data[0x1B] & 0x0F) == 0)
				printf2("Unknown");
			else
				printf2("%u", data[0x1B] & 0x0F);
			printf2("\n");
			break;

		case 18: /* 3.3.19 32-bit Memory Error Information */
			printf2("32-bit Memory Error Information\n");
			if (h->length < 0x17) break;
			printf2("\tType: %s\n",
				dmi_memory_error_type(data[0x04]));
			printf2("\tGranularity: %s\n",
				dmi_memory_error_granularity(data[0x05]));
			printf2("\tOperation: %s\n",
				dmi_memory_error_operation(data[0x06]));
			printf2("\tVendor Syndrome:");
			dmi_memory_error_syndrome(DWORD(data + 0x07));
			printf2("\n");
			printf2("\tMemory Array Address:");
			dmi_32bit_memory_error_address(DWORD(data + 0x0B));
			printf2("\n");
			printf2("\tDevice Address:");
			dmi_32bit_memory_error_address(DWORD(data + 0x0F));
			printf2("\n");
			printf2("\tResolution:");
			dmi_32bit_memory_error_address(DWORD(data + 0x13));
			printf2("\n");
			break;

		case 19: /* 3.3.20 Memory Array Mapped Address */
			printf2("Memory Array Mapped Address\n");
			if (h->length < 0x0F) break;
			printf2("\tStarting Address: 0x%08X%03X\n",
				DWORD(data + 0x04) >> 2, (DWORD(data + 0x04) & 0x3) << 10);
			printf2("\tEnding Address: 0x%08X%03X\n",
				DWORD(data + 0x08) >> 2, ((DWORD(data + 0x08) & 0x3) << 10) + 0x3FF);
			printf2("\tRange Size:");
			dmi_mapped_address_size(DWORD(data + 0x08) - DWORD(data + 0x04) + 1);
			printf2("\n");
			if (!(opt.flags & FLAG_QUIET))
				printf2("\tPhysical Array Handle: 0x%04X\n",
					WORD(data + 0x0C));
			printf2("\tPartition Width: %u\n",
				data[0x0F]);
			break;

		case 20: /* 3.3.21 Memory Device Mapped Address */
			printf2("Memory Device Mapped Address\n");
			if (h->length < 0x13) break;
			printf2("\tStarting Address: 0x%08X%03X\n",
				DWORD(data + 0x04) >> 2, (DWORD(data + 0x04) & 0x3) << 10);
			printf2("\tEnding Address: 0x%08X%03X\n",
				DWORD(data + 0x08) >> 2, ((DWORD(data + 0x08) & 0x3) << 10) + 0x3FF);
			printf2("\tRange Size:");
			dmi_mapped_address_size(DWORD(data + 0x08) - DWORD(data + 0x04) + 1);
			printf2("\n");
			if (!(opt.flags & FLAG_QUIET))
			{
				printf2("\tPhysical Device Handle: 0x%04X\n",
					WORD(data + 0x0C));
				printf2("\tMemory Array Mapped Address Handle: 0x%04X\n",
					WORD(data + 0x0E));
			}
			printf2("\tPartition Row Position:");
			dmi_mapped_address_row_position(data[0x10]);
			printf2("\n");
			dmi_mapped_address_interleave_position(data[0x11], "\t");
			dmi_mapped_address_interleaved_data_depth(data[0x12], "\t");
			break;

		case 21: /* 3.3.22 Built-in Pointing Device */
			printf2("Built-in Pointing Device\n");
			if (h->length < 0x07) break;
			printf2("\tType: %s\n",
				dmi_pointing_device_type(data[0x04]));
			printf2("\tInterface: %s\n",
				dmi_pointing_device_interface(data[0x05]));
			printf2("\tButtons: %u\n",
				data[0x06]);
			break;

		case 22: /* 3.3.23 Portable Battery */
			printf2("Portable Battery\n");
			if (h->length < 0x10) break;
			printf2("\tLocation: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x05]));
			if (data[0x06] || h->length < 0x1A)
				printf2("\tManufacture Date: %s\n",
					dmi_string(h, data[0x06]));
			if (data[0x07] || h->length < 0x1A)
				printf2("\tSerial Number: %s\n",
					dmi_string(h, data[0x07]));
			printf2("\tName: %s\n",
				dmi_string(h, data[0x08]));
			if (data[0x09] != 0x02 || h->length < 0x1A)
				printf2("\tChemistry: %s\n",
					dmi_battery_chemistry(data[0x09]));
			printf2("\tDesign Capacity:");
			if (h->length < 0x1A)
				dmi_battery_capacity(WORD(data + 0x0A), 1);
			else
				dmi_battery_capacity(WORD(data + 0x0A), data[0x15]);
			printf2("\n");
			printf2("\tDesign Voltage:");
			dmi_battery_voltage(WORD(data + 0x0C));
			printf2("\n");
			printf2("\tSBDS Version: %s\n",
				dmi_string(h, data[0x0E]));
			printf2("\tMaximum Error:");
			dmi_battery_maximum_error(data[0x0F]);
			printf2("\n");
			if (h->length < 0x1A) break;
			if (data[0x07] == 0)
				printf2("\tSBDS Serial Number: %04X\n",
					WORD(data + 0x10));
			if (data[0x06] == 0)
				printf2("\tSBDS Manufacture Date: %u-%02u-%02u\n",
					1980 + (WORD(data + 0x12) >> 9),
					(WORD(data + 0x12) >> 5) & 0x0F,
					WORD(data + 0x12) & 0x1F);
			if (data[0x09] == 0x02)
				printf2("\tSBDS Chemistry: %s\n",
					dmi_string(h, data[0x14]));
			printf2("\tOEM-specific Information: 0x%08X\n",
				DWORD(data + 0x16));
			break;

		case 23: /* 3.3.24 System Reset */
			printf2("System Reset\n");
			if (h->length < 0x0D) break;
			printf2("\tStatus: %s\n",
				data[0x04] & (1 << 0) ? "Enabled" : "Disabled");
			printf2("\tWatchdog Timer: %s\n",
				data[0x04] & (1 << 5) ? "Present" : "Not Present");
			if (!(data[0x04] & (1 << 5)))
				break;
			printf2("\tBoot Option: %s\n",
				dmi_system_reset_boot_option((data[0x04] >> 1) & 0x3));
			printf2("\tBoot Option On Limit: %s\n",
				dmi_system_reset_boot_option((data[0x04] >> 3) & 0x3));
			printf2("\tReset Count:");
			dmi_system_reset_count(WORD(data + 0x05));
			printf2("\n");
			printf2("\tReset Limit:");
			dmi_system_reset_count(WORD(data + 0x07));
			printf2("\n");
			printf2("\tTimer Interval:");
			dmi_system_reset_timer(WORD(data + 0x09));
			printf2("\n");
			printf2("\tTimeout:");
			dmi_system_reset_timer(WORD(data + 0x0B));
			printf2("\n");
			break;

		case 24: /* 3.3.25 Hardware Security */
			printf2("Hardware Security\n");
			if (h->length < 0x05) break;
			printf2("\tPower-On Password Status: %s\n",
				dmi_hardware_security_status(data[0x04] >> 6));
			printf2("\tKeyboard Password Status: %s\n",
				dmi_hardware_security_status((data[0x04] >> 4) & 0x3));
			printf2("\tAdministrator Password Status: %s\n",
				dmi_hardware_security_status((data[0x04] >> 2) & 0x3));
			printf2("\tFront Panel Reset Status: %s\n",
				dmi_hardware_security_status(data[0x04] & 0x3));
			break;

		case 25: /* 3.3.26 System Power Controls */
			printf2("\tSystem Power Controls\n");
			if (h->length < 0x09) break;
			printf2("\tNext Scheduled Power-on:");
			dmi_power_controls_power_on(data + 0x04);
			printf2("\n");
			break;

		case 26: /* 3.3.27 Voltage Probe */
			printf2("Voltage Probe\n");
			if (h->length < 0x14) break;
			printf2("\tDescription: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tLocation: %s\n",
				dmi_voltage_probe_location(data[0x05] & 0x1f));
			printf2("\tStatus: %s\n",
				dmi_probe_status(data[0x05] >> 5));
			printf2("\tMaximum Value:");
			dmi_voltage_probe_value(WORD(data + 0x06));
			printf2("\n");
			printf2("\tMinimum Value:");
			dmi_voltage_probe_value(WORD(data + 0x08));
			printf2("\n");
			printf2("\tResolution:");
			dmi_voltage_probe_resolution(WORD(data + 0x0A));
			printf2("\n");
			printf2("\tTolerance:");
			dmi_voltage_probe_value(WORD(data + 0x0C));
			printf2("\n");
			printf2("\tAccuracy:");
			dmi_probe_accuracy(WORD(data + 0x0E));
			printf2("\n");
			printf2("\tOEM-specific Information: 0x%08X\n",
				DWORD(data + 0x10));
			if (h->length < 0x16) break;
			printf2("\tNominal Value:");
			dmi_voltage_probe_value(WORD(data + 0x14));
			printf2("\n");
			break;

		case 27: /* 3.3.28 Cooling Device */
			printf2("Cooling Device\n");
			if (h->length < 0x0C) break;
			if (!(opt.flags & FLAG_QUIET) && WORD(data + 0x04) != 0xFFFF)
				printf2("\tTemperature Probe Handle: 0x%04X\n",
					WORD(data + 0x04));
			printf2("\tType: %s\n",
				dmi_cooling_device_type(data[0x06] & 0x1f));
			printf2("\tStatus: %s\n",
				dmi_probe_status(data[0x06] >> 5));
			if (data[0x07] != 0x00)
				printf2("\tCooling Unit Group: %u\n",
					data[0x07]);
			printf2("\tOEM-specific Information: 0x%08X\n",
				DWORD(data + 0x08));
			if (h->length < 0x0E) break;
			printf2("\tNominal Speed:");
			dmi_cooling_device_speed(WORD(data + 0x0C));
			printf2("\n");
			break;

		case 28: /* 3.3.29 Temperature Probe */
			printf2("Temperature Probe\n");
			if (h->length < 0x14) break;
			printf2("\tDescription: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tLocation: %s\n",
				dmi_temperature_probe_location(data[0x05] & 0x1F));
			printf2("\tStatus: %s\n",
				dmi_probe_status(data[0x05] >> 5));
			printf2("\tMaximum Value:");
			dmi_temperature_probe_value(WORD(data + 0x06));
			printf2("\n");
			printf2("\tMinimum Value:");
			dmi_temperature_probe_value(WORD(data + 0x08));
			printf2("\n");
			printf2("\tResolution:");
			dmi_temperature_probe_resolution(WORD(data + 0x0A));
			printf2("\n");
			printf2("\tTolerance:");
			dmi_temperature_probe_value(WORD(data + 0x0C));
			printf2("\n");
			printf2("\tAccuracy:");
			dmi_probe_accuracy(WORD(data + 0x0E));
			printf2("\n");
			printf2("\tOEM-specific Information: 0x%08X\n",
				DWORD(data + 0x10));
			if (h->length < 0x16) break;
			printf2("\tNominal Value:");
			dmi_temperature_probe_value(WORD(data + 0x14));
			printf2("\n");
			break;

		case 29: /* 3.3.30 Electrical Current Probe */
			printf2("Electrical Current Probe\n");
			if (h->length < 0x14) break;
			printf2("\tDescription: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tLocation: %s\n",
				dmi_voltage_probe_location(data[5] & 0x1F));
			printf2("\tStatus: %s\n",
				dmi_probe_status(data[0x05] >> 5));
			printf2("\tMaximum Value:");
			dmi_current_probe_value(WORD(data + 0x06));
			printf2("\n");
			printf2("\tMinimum Value:");
			dmi_current_probe_value(WORD(data + 0x08));
			printf2("\n");
			printf2("\tResolution:");
			dmi_current_probe_resolution(WORD(data + 0x0A));
			printf2("\n");
			printf2("\tTolerance:");
			dmi_current_probe_value(WORD(data + 0x0C));
			printf2("\n");
			printf2("\tAccuracy:");
			dmi_probe_accuracy(WORD(data + 0x0E));
			printf2("\n");
			printf2("\tOEM-specific Information: 0x%08X\n",
				DWORD(data + 0x10));
			if (h->length < 0x16) break;
			printf2("\tNominal Value:");
			dmi_current_probe_value(WORD(data + 0x14));
			printf2("\n");
			break;

		case 30: /* 3.3.31 Out-of-band Remote Access */
			printf2("Out-of-band Remote Access\n");
			if (h->length < 0x06) break;
			printf2("\tManufacturer Name: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tInbound Connection: %s\n",
				data[0x05] & (1 << 0) ? "Enabled" : "Disabled");
			printf2("\tOutbound Connection: %s\n",
				data[0x05] & (1 << 1) ? "Enabled" : "Disabled");
			break;

		case 31: /* 3.3.32 Boot Integrity Services Entry Point */
			printf2("Boot Integrity Services Entry Point\n");
			if (h->length < 0x1C) break;
			printf2("\tChecksum: %s\n",
				checksum(data, h->length) ? "OK" : "Invalid");
			printf2("\t16-bit Entry Point Address: %04X:%04X\n",
				DWORD(data + 0x08) >> 16,
				DWORD(data + 0x08) & 0xFFFF);
			printf2("\t32-bit Entry Point Address: 0x%08X\n",
				DWORD(data + 0x0C));
			break;

		case 32: /* 3.3.33 System Boot Information */
			printf2("System Boot Information\n");
			if (h->length < 0x0B) break;
			printf2("\tStatus: %s\n",
				dmi_system_boot_status(data[0x0A]));
			break;

		case 33: /* 3.3.34 64-bit Memory Error Information */
			if (h->length < 0x1F) break;
			printf2("64-bit Memory Error Information\n");
			printf2("\tType: %s\n",
				dmi_memory_error_type(data[0x04]));
			printf2("\tGranularity: %s\n",
				dmi_memory_error_granularity(data[0x05]));
			printf2("\tOperation: %s\n",
				dmi_memory_error_operation(data[0x06]));
			printf2("\tVendor Syndrome:");
			dmi_memory_error_syndrome(DWORD(data + 0x07));
			printf2("\n");
			printf2("\tMemory Array Address:");
			dmi_64bit_memory_error_address(QWORD(data + 0x0B));
			printf2("\n");
			printf2("\tDevice Address:");
			dmi_64bit_memory_error_address(QWORD(data + 0x13));
			printf2("\n");
			printf2("\tResolution:");
			dmi_32bit_memory_error_address(DWORD(data + 0x1B));
			printf2("\n");
			break;

		case 34: /* 3.3.35 Management Device */
			printf2("Management Device\n");
			if (h->length < 0x0B) break;
			printf2("\tDescription: %s\n",
				dmi_string(h, data[0x04]));
			printf2("\tType: %s\n",
				dmi_management_device_type(data[0x05]));
			printf2("\tAddress: 0x%08X\n",
				DWORD(data + 0x06));
			printf2("\tAddress Type: %s\n",
				dmi_management_device_address_type(data[0x0A]));
			break;

		case 35: /* 3.3.36 Management Device Component */
			printf2("Management Device Component\n");
			if (h->length < 0x0B) break;
			printf2("\tDescription: %s\n",
				dmi_string(h, data[0x04]));
			if (!(opt.flags & FLAG_QUIET))
			{
				printf2("\tManagement Device Handle: 0x%04X\n",
					WORD(data + 0x05));
				printf2("\tComponent Handle: 0x%04X\n",
					WORD(data + 0x07));
				if (WORD(data + 0x09) != 0xFFFF)
					printf2("\tThreshold Handle: 0x%04X\n",
					WORD(data + 0x09));
			}
			break;

		case 36: /* 3.3.37 Management Device Threshold Data */
			printf2("Management Device Threshold Data\n");
			if (h->length < 0x10) break;
			if (WORD(data + 0x04) != 0x8000)
				printf2("\tLower Non-critical Threshold: %d\n",
					(i16)WORD(data + 0x04));
			if (WORD(data + 0x06) != 0x8000)
				printf2("\tUpper Non-critical Threshold: %d\n",
					(i16)WORD(data + 0x06));
			if (WORD(data + 0x08) != 0x8000)
				printf2("\tLower Critical Threshold: %d\n",
					(i16)WORD(data + 0x08));
			if (WORD(data + 0x0A) != 0x8000)
				printf2("\tUpper Critical Threshold: %d\n",
					(i16)WORD(data + 0x0A));
			if (WORD(data + 0x0C) != 0x8000)
				printf2("\tLower Non-recoverable Threshold: %d\n",
					(i16)WORD(data + 0x0C));
			if (WORD(data + 0x0E) != 0x8000)
				printf2("\tUpper Non-recoverable Threshold: %d\n",
					(i16)WORD(data + 0x0E));
			break;

		case 37: /* 3.3.38 Memory Channel */
			printf2("Memory Channel\n");
			if (h->length < 0x07) break;
			printf2("\tType: %s\n",
				dmi_memory_channel_type(data[0x04]));
			printf2("\tMaximal Load: %u\n",
				data[0x05]);
			printf2("\tDevices: %u\n",
				data[0x06]);
			if (h->length < 0x07 + 3 * data[0x06]) break;
			dmi_memory_channel_devices(data[0x06], data + 0x07, "\t");
			break;

		case 38: /* 3.3.39 IPMI Device Information */
			/*
			 * We use the word "Version" instead of "Revision", conforming to
			 * the IPMI specification.
			 */
			printf2("IPMI Device Information\n");
			if (h->length < 0x10) break;
			printf2("\tInterface Type: %s\n",
				dmi_ipmi_interface_type(data[0x04]));
			printf2("\tSpecification Version: %u.%u\n",
				data[0x05] >> 4, data[0x05] & 0x0F);
			printf2("\tI2C Slave Address: 0x%02x\n",
				data[0x06] >> 1);
			if (data[0x07] != 0xFF)
				printf2("\tNV Storage Device Address: %u\n",
					data[0x07]);
			else
				printf2("\tNV Storage Device: Not Present\n");
			printf2("\tBase Address: ");
			dmi_ipmi_base_address(data[0x04], data + 0x08,
				h->length < 0x12 ? 0 : (data[0x10] >> 5) & 1);
			printf2("\n");
			if (h->length < 0x12) break;
			if (data[0x04] != 0x04)
			{
				printf2("\tRegister Spacing: %s\n",
					dmi_ipmi_register_spacing(data[0x10] >> 6));
				if (data[0x10] & (1 << 3))
				{
					printf2("\tInterrupt Polarity: %s\n",
						data[0x10] & (1 << 1) ? "Active High" : "Active Low");
					printf2("\tInterrupt Trigger Mode: %s\n",
						data[0x10] & (1 << 0) ? "Level" : "Edge");
				}
			}
			if (data[0x11] != 0x00)
			{
				printf2("\tInterrupt Number: %x\n",
					data[0x11]);
			}
			break;

		case 39: /* 3.3.40 System Power Supply */
			printf2("System Power Supply\n");
			if (h->length < 0x10) break;
			if (data[0x04] != 0x00)
				printf2("\tPower Unit Group: %u\n",
					data[0x04]);
			printf2("\tLocation: %s\n",
				dmi_string(h, data[0x05]));
			printf2("\tName: %s\n",
				dmi_string(h, data[0x06]));
			printf2("\tManufacturer: %s\n",
				dmi_string(h, data[0x07]));
			printf2("\tSerial Number: %s\n",
				dmi_string(h, data[0x08]));
			printf2("\tAsset Tag: %s\n",
				dmi_string(h, data[0x09]));
			printf2("\tModel Part Number: %s\n",
				dmi_string(h, data[0x0A]));
			printf2("\tRevision: %s\n",
				dmi_string(h, data[0x0B]));
			printf2("\tMax Power Capacity:");
			dmi_power_supply_power(WORD(data + 0x0C));
			printf2("\n");
			printf2("\tStatus:");
			if (WORD(data + 0x0E) & (1 << 1))
				printf2(" Present, %s",
					dmi_power_supply_status((WORD(data + 0x0E) >> 7) & 0x07));
			else
				printf2(" Not Present");
			printf2("\n");
			printf2("\tType: %s\n",
				dmi_power_supply_type((WORD(data + 0x0E) >> 10) & 0x0F));
			printf2("\tInput Voltage Range Switching: %s\n",
				dmi_power_supply_range_switching((WORD(data + 0x0E) >> 3) & 0x0F));
			printf2("\tPlugged: %s\n",
				WORD(data + 0x0E) & (1 << 2) ? "No" : "Yes");
			printf2("\tHot Replaceable: %s\n",
				WORD(data + 0x0E) & (1 << 0) ? "Yes" : "No");
			if (h->length < 0x16) break;
			if (!(opt.flags & FLAG_QUIET))
			{
				if (WORD(data + 0x10) != 0xFFFF)
					printf2("\tInput Voltage Probe Handle: 0x%04X\n",
						WORD(data + 0x10));
				if (WORD(data + 0x12) != 0xFFFF)
					printf2("\tCooling Device Handle: 0x%04X\n",
						WORD(data + 0x12));
				if (WORD(data + 0x14) != 0xFFFF)
					printf2("\tInput Current Probe Handle: 0x%04X\n",
						WORD(data + 0x14));
			}
			break;

		case 40: /* 3.3.41 Additional Information */
			if (h->length < 0x0B) break;
			if (!(opt.flags & FLAG_QUIET))
				dmi_additional_info(h, "");
			break;

		case 41: /* 3.3.42 Onboard Device Extended Information */
			printf2("Onboard Device\n");
			if (h->length < 0x0B) break;
			printf2("\tReference Designation: %s\n", dmi_string(h, data[0x04]));
			printf2("\tType: %s\n",
				dmi_on_board_devices_type(data[0x05] & 0x7F));
			printf2("\tStatus: %s\n",
				data[0x05] & 0x80 ? "Enabled" : "Disabled");
			printf2("\tType Instance: %u\n", data[0x06]);
			dmi_slot_segment_bus_func(WORD(data + 0x07), data[0x09], data[0x0A], "\t");
			break;

		case 126: /* 3.3.43 Inactive */
			printf2("Inactive\n");
			break;

		case 127: /* 3.3.44 End Of Table */
			printf2("End Of Table\n");
			break;

		default:
			if (dmi_decode_oem(h))
				break;
			if (opt.flags & FLAG_QUIET)
				return;
			printf2("%s Type\n",
				h->type >= 128 ? "OEM-specific" : "Unknown");
			dmi_dump(h, "\t");
	}
	printf2("\n");
}

static void to_dmi_header(struct dmi_header *h, u8 *data)
{
	h->type = data[0];
	h->length = data[1];
	h->handle = WORD(data + 2);
	h->data = data;
}

static void dmi_table_string(const struct dmi_header *h, const u8 *data, u16 ver)
{
	int key;
	u8 offset = opt.string->offset;

	if (offset >= h->length)
		return;

	key = (opt.string->type << 8) | offset;
	switch (key)
	{
		case 0x108:
			dmi_system_uuid(data + offset, ver);
			printf2("\n");
			break;
		case 0x305:
			printf2("%s\n", dmi_chassis_type(data[offset]));
			break;
		case 0x406:
			printf2("%s\n", dmi_processor_family(h));
			break;
		case 0x416:
			dmi_processor_frequency(data + offset);
			printf2("\n");
			break;
		default:
			printf2("%s\n", dmi_string(h, data[offset]));
	}
}

static void dmi_table_dump(u32 base, u16 len, const char *devmem)
{
	u8 *buf;

	if ((buf = mem_chunk(base, len, devmem)) == NULL)
	{
		fprintf2(stderr, "Failed to read table, sorry.\n");
		return;
	}

	if (!(opt.flags & FLAG_QUIET))
		printf2("# Writing %d bytes to %s.\n", len, opt.dumpfile);
	write_dump(32, len, buf, opt.dumpfile, 0);
	free(buf);
}

static void dmi_table(u32 base, u16 len, u16 num, u16 ver, const char *devmem)
{
	
	u8 *buf;
	u8 *data;
	int i = 0;

	if (opt.flags & FLAG_DUMP_BIN)
	{
		dmi_table_dump(base, len, devmem);
		return;
	}

	if (!(opt.flags & FLAG_QUIET))
	{
		if (opt.type == NULL)
		{
			printf2("%u structures occupying %u bytes.\n",
				num, len);
			if (!(opt.flags & FLAG_FROM_DUMP))
				printf2("Table at 0x%08X.\n", base);
		}
		printf2("\n");
	}
	/*
	 * if devmem is NULL then base has the SMBIOS Table address
	 * already allocated and not the physical memory address that 
	 * needs to be mapped.
	 * This change was made to support Windows 2003 that blocks
	 * access to physical memory but returns the SMBIOS Table
	 * througth GetSystemFirmwareTable API.
	 *
	 * see more on winsmbios.h and winsmbios.c
	 */	
	if (devmem == NULL) 
	{
    buf = (u8 *)base;
  } else {
	if ((buf = mem_chunk(base, len, devmem)) == NULL)
	{
		fprintf2(stderr, "Table is unreachable, sorry."
#ifndef USE_MMAP
			" Try compiling dmidecode with -DUSE_MMAP."
#endif
			"\n");
		return;
		}
	}
	data = buf;
	while (i < num && data+4 <= buf + len) /* 4 is the length of an SMBIOS structure header */
	{
		u8 *next;
		struct dmi_header h;
		int display;

		to_dmi_header(&h, data);
		display = ((opt.type == NULL || opt.type[h.type])
			&& !((opt.flags & FLAG_QUIET) && (h.type > 39 && h.type <= 127))
			&& !opt.string);

		/*
		 * If a short entry is found (less than 4 bytes), not only it
		 * is invalid, but we cannot reliably locate the next entry.
		 * Better stop at this point, and let the user know his/her
		 * table is broken.
		 */
		if (h.length < 4)
		{
			printf2("Invalid entry length (%u). DMI table is "
			       "broken! Stop.\n\n", (unsigned int)h.length);
			opt.flags |= FLAG_QUIET;
			break;
		}

		/* In quiet mode, stop decoding at end of table marker */
		if ((opt.flags & FLAG_QUIET) && h.type == 127)
			break;

		if (display
		 && (!(opt.flags & FLAG_QUIET) || (opt.flags & FLAG_DUMP)))
			printf2("Handle 0x%04X, DMI type %d, %d bytes\n",
				h.handle, h.type, h.length);

		/* assign vendor for vendor-specific decodes later */
		if (h.type == 0 && h.length >= 5)
			dmi_set_vendor(dmi_string(&h, data[0x04]));

		/* look for the next handle */
		next = data + h.length;
		while (next - buf + 1 < len && (next[0] != 0 || next[1] != 0))
			next++;
		next += 2;
		if (display)
		{
			if (next - buf <= len)
			{
				if (opt.flags & FLAG_DUMP)
				{
					dmi_dump(&h, "\t");
					printf2("\n");
				}
				else
					dmi_decode(&h, ver);
			}
			else if (!(opt.flags & FLAG_QUIET))
				printf2("\t<TRUNCATED>\n\n");
		}
		else if (opt.string != NULL
		      && opt.string->type == h.type)
			dmi_table_string(&h, data, ver);

		data = next;
		i++;
	}

	if (!(opt.flags & FLAG_QUIET))
	{
		if (i != num)
			printf2("Wrong DMI structures count: %d announced, "
				"only %d decoded.\n", num, i);
		if (data - buf != len)
			printf2("Wrong DMI structures length: %d bytes "
				"announced, structures occupy %d bytes.\n",
				len, (unsigned int)(data - buf));
	}

	free(buf);
}


/*
 * Build a crafted entry point with table address hard-coded to 32,
 * as this is where we will put it in the output file. We adjust the
 * DMI checksum appropriately. The SMBIOS checksum needs no adjustment.
 */
static void overwrite_dmi_address(u8 *buf)
{
	buf[0x05] += buf[0x08] + buf[0x09] + buf[0x0A] + buf[0x0B] - 32;
	buf[0x08] = 32;
	buf[0x09] = 0;
	buf[0x0A] = 0;
	buf[0x0B] = 0;
}

static int smbios_decode(u8 *buf, const char *devmem)
{
	u16 ver;

	if (!checksum(buf, buf[0x05])
	 || memcmp(buf + 0x10, "_DMI_", 5) != 0
	 || !checksum(buf + 0x10, 0x0F))
		return 0;

	ver = (buf[0x06] << 8) + buf[0x07];
	/* Some BIOS report weird SMBIOS version, fix that up */
	switch (ver)
	{
		case 0x021F:
			if (!(opt.flags & FLAG_QUIET))
				printf2("SMBIOS version fixup (2.%d -> 2.%d).\n",
				       31, 3);
			ver = 0x0203;
			break;
		case 0x0233:
			if (!(opt.flags & FLAG_QUIET))
				printf2("SMBIOS version fixup (2.%d -> 2.%d).\n",
				       51, 6);
			ver = 0x0206;
			break;
	}
	if (!(opt.flags & FLAG_QUIET))
		printf2("SMBIOS %u.%u present.\n",
			ver >> 8, ver & 0xFF);

	dmi_table(DWORD(buf + 0x18), WORD(buf + 0x16), WORD(buf + 0x1C),
		ver, devmem);

	if (opt.flags & FLAG_DUMP_BIN)
	{
		u8 crafted[32];

		memcpy(crafted, buf, 32);
		overwrite_dmi_address(crafted + 0x10);

		if (!(opt.flags & FLAG_QUIET))
			printf2("# Writing %d bytes to %s.\n", crafted[0x05],
				opt.dumpfile);
		write_dump(0, crafted[0x05], crafted, opt.dumpfile, 1);
	}

	return 1;
}

static int legacy_decode(u8 *buf, const char *devmem)
{
	if (!checksum(buf, 0x0F))
		return 0;

	if (!(opt.flags & FLAG_QUIET))
		printf2("Legacy DMI %u.%u present.\n",
			buf[0x0E] >> 4, buf[0x0E] & 0x0F);

	dmi_table(DWORD(buf + 0x08), WORD(buf + 0x06), WORD(buf + 0x0C),
		((buf[0x0E] & 0xF0) << 4) + (buf[0x0E] & 0x0F), devmem);

	if (opt.flags & FLAG_DUMP_BIN)
	{
		u8 crafted[16];

		memcpy(crafted, buf, 16);
		overwrite_dmi_address(crafted);

		printf2("# Writing %d bytes to %s.\n", 0x0F, opt.dumpfile);
		write_dump(0, 0x0F, crafted, opt.dumpfile, 1);
	}

	return 1;
}

/*
 * Probe for EFI interface
 */
#define EFI_NOT_FOUND   (-1)
#define EFI_NO_SMBIOS   (-2)
static int address_from_efi(size_t *address)
{
	FILE *efi_systab;
	const char *filename;
	char linebuf[64];
	int ret;

	*address = 0; /* Prevent compiler warning */

	/*
	 * Linux up to 2.6.6: /proc/efi/systab
	 * Linux 2.6.7 and up: /sys/firmware/efi/systab
	 */
	if ((efi_systab = fopen(filename = "/sys/firmware/efi/systab", "r")) == NULL
	 && (efi_systab = fopen(filename = "/proc/efi/systab", "r")) == NULL)
	{
		/* No EFI interface, fallback to memory scan */
		return EFI_NOT_FOUND;
	}
	ret = EFI_NO_SMBIOS;
	while ((fgets(linebuf, sizeof(linebuf) - 1, efi_systab)) != NULL)
	{
		char *addrp = strchr(linebuf, '=');
		*(addrp++) = '\0';
		if (strcmp(linebuf, "SMBIOS") == 0)
		{
			*address = strtoul(addrp, NULL, 0);
			if (!(opt.flags & FLAG_QUIET))
				printf2("# SMBIOS entry point at 0x%08lx\n",
				       (unsigned long)*address);
			ret = 0;
			break;
		}
	}
	if (fclose(efi_systab) != 0)
		perror(filename);

	if (ret == EFI_NO_SMBIOS)
		fprintf2(stderr, "%s: SMBIOS entry point missing\n", filename);
	return ret;
}

//int main(int argc, char * const argv[])
int get_dev_info(BufferGetter bgetter)
{
	int argc = 1;
	char * const argv[] = { "dmidecode" };

	int ret = 0;                /* Returned value */
	int found = 0;
	size_t fp;
	int efi;
	u8 *buf;
	u8 *dmibuf = 0;
	size_t dmilen = 0;

  if (!bgetter) 
      bgetter = print_buf;

#if defined(_WIN32) || defined(_WIN64)
    /*
     * these varibles are used only when run on windows 2003 or above.
     * Since these versions block access to physical memory.
     * Windows NT, 2000 and XP still accessing physical memory througth
     * mem_chunck
     */
    int num_structures = 0;
    PRawSMBIOSData smb = NULL;
#endif /* _WIN32 */
	
	if (sizeof(u8) != 1 || sizeof(u16) != 2 || sizeof(u32) != 4 || '\0' != 0)
	{
		fprintf(stderr, "%s: compiler incompatibility\n", argv[0]);
		// exit(255);
	}

	/* Set default option values */
	opt.devmem = DEFAULT_MEM_DEV;
	opt.flags = 0;


	// if (parse_command_line(argc, argv)<0)
	// {
	// 	ret = 2;
	// 	goto exit_free;
	// }

	if (opt.flags & FLAG_HELP)
	{
		// print_help();
		goto exit_free;
	}

	if (opt.flags & FLAG_VERSION)
	{
		printf2("%s\n", VERSION);
		goto exit_free;
	}

	if (!(opt.flags & FLAG_QUIET))
		printf2("# dmidecode %s\n", VERSION);

	/* Read from dump if so instructed */
	if (opt.flags & FLAG_FROM_DUMP)
	{
		if (!(opt.flags & FLAG_QUIET))
			printf2("Reading SMBIOS/DMI data from file %s.\n",
			       opt.dumpfile);
		if ((buf = mem_chunk(0, 0x20, opt.dumpfile)) == NULL)
		{
			ret = 1;
			goto exit_free;
		}

		if (memcmp(buf, "_SM_", 4) == 0)
		{
			if (smbios_decode(buf, opt.dumpfile))
				found++;
		}
		else if (memcmp(buf, "_DMI_", 5) == 0)
		{
			if (legacy_decode(buf, opt.dumpfile))
				found++;
		}
		goto done;
	}

	/* First try EFI (ia64, Intel-based Mac) */
	efi = address_from_efi(&fp);
	switch (efi)
	{
		case EFI_NOT_FOUND:
			goto memory_scan;
		case EFI_NO_SMBIOS:
			ret = 1;
			goto exit_free;
	}

	if ((buf = mem_chunk(fp, 0x20, opt.devmem)) == NULL)
	{
		ret = 1;
		goto exit_free;
	}

	if (smbios_decode(buf, opt.devmem))
		found++;
	goto done;

memory_scan:
	/* Fallback to memory scan (x86, x86_64) */
		
    /*
     * If running on windows, checks if its Windows 2003 or vista and
     * get the SMBIOS data without access to physical memory.
     * If its Windows NT, 2000 or XP, access the physical memory and
     * scans for SMBIOS table entry point, just like all other OS.
     * If its Windows 9x or Me, print error and exits.
     */
#if defined(_WIN32) || defined(_WIN64) //def __WIN32__
    switch(get_windows_platform()){
                                   
        case WIN_2003_VISTA://gets the SMBIOS table, prints values and exits

            //loads the GetSystemFirmwareTable function
            if(!LocateNtdllEntryPoints()){
       	        return printf2("erro ao carregar a dll");
                goto exit_free;             	
            }

            //gets the raw smbios table
            smb = get_raw_smbios_table();
            num_structures = count_smbios_structures(&smb->SMBIOSTableData[0], smb->Length);

    		if(!(opt.flags & FLAG_QUIET)){
    			printf2("SMBIOS %u.%u present.\n", smb->SMBIOSMajorVersion, 
                    smb->SMBIOSMinorVersion);
            }
            
            //shows the smbios information
	          // dmi_table((u32)&smb->SMBIOSTableData[0], smb->Length, num_structures,
    	      // 		(smb->SMBIOSMajorVersion<<8)+smb->SMBIOSMinorVersion, NULL);   
            
            bgetter(&smb->SMBIOSTableData[0], smb->Length);
            // printf("load smb use SMBIOSTableData v%d.%d\n", smb->SMBIOSMajorVersion, smb->SMBIOSMinorVersion);
            free(smb);
            goto exit_free;
        break;
        
        case WIN_UNSUPORTED://prints error and exits
            printf2("\nDMIDECODE runs on Windows NT/2000/XP. Windows 9x/Me are not supported.\n");
            goto exit_free;
        break;
        
        default:
            /*
             * do nothing. Follow the code below, scans for the 
             * SMBIOS table entry point, etc...
             */
        break;
    }
#endif  /*__WIN32__*/

	if ((buf = mem_chunk(0xF0000, 0x10000, opt.devmem)) == NULL)
	{
		ret = 1;
		goto exit_free;
	}

	for (fp = 0; fp <= 0xFFF0; fp += 16)
	{
		if (memcmp(buf + fp, "_SM_", 4) == 0 && fp <= 0xFFE0)
		{
			if (smbios_decode(buf+fp, opt.devmem)) {
				found++;
				// fp += 16; // addr to '_DMI_'

				u8 * addr = DWORD(fp + buf + 0x18);
				dmilen = WORD(fp + buf + 0x16);
				dmibuf = mem_chunk(addr, dmilen, opt.devmem);
				// printf("found smbios header: %x, dmi address: %x, len: %d\n", 
				// 			 buf + fp, addr, dmilen);
			}
		}
		else if (memcmp(buf + fp, "_DMI_", 5) == 0)
		{
			if (legacy_decode(buf + fp, opt.devmem)) {
				found++;

				u8 * addr = DWORD(fp + buf + 0x08);
				dmilen = WORD(fp + buf + 0x06);
				dmibuf = mem_chunk(addr, dmilen, opt.devmem);
				// printf("found DMI legacy \n");
			}
		}
	}

done:
	if (!found) {
		printf2("# No SMBIOS nor DMI entry point found, sorry.\n");
    exit(1);
  }

	if (dmibuf) {
		bgetter(dmibuf, dmilen);
		free(dmibuf);
	}

	free(buf);

exit_free:
	free(opt.type);
	return ret;
}


void print_buf(u8 *buf, u16 len) {
	int i = 0, j = 0;
	int s = 0, line = 16;
	printf("Buf len %d\n", len);

	for (i=0; i<len; ++i) {
		printf("%02x ", buf[i]);

		if (i % line == line-1) {
			printf(" | ");
			s = i;
			for (j=s-line; j<s; ++j) {
				if (buf[j] == '\n' || buf[j] == ' ' || buf[j] == '\r') {
					printf(".");
				} else {
					printf("%c", buf[j]);
				}
			}
			printf("\n");
		}
	}
}